package debug

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/eth"
)

var transferTracer = `
{
  reverted: false,
  transfers: [],

  // fault is invoked when the actual execution of an opcode fails.
  fault(log, db) {
    this.reverted = true;
  },

  // step is invoked for every opcode that the VM executes.
  step(log, db) {
    // Capture any errors immediately
    const error = log.getError();
    if (error !== undefined) {
      this.fault(log, db);
    } else {
      const op = log.op.toString();
      switch (op) {
        case 'CALL':
        case 'CALLCODE':
        case 'DELEGATECALL':
          this.handleCall(log, op);
          break;

        case 'REVERT':
          this.reverted = true;
          break;
      }
    }
  },

  handleCall(log, op) {
    const to = toAddress(log.stack.peek(1).toString(16));
    if (!isPrecompiled(to)) {
      if (op != 'DELEGATECALL') {
        valueBigInt = bigInt(log.stack.peek(2));
        if (valueBigInt.gt(0)) {
          const transfer = {
            type: 'cGLD nested transfer',
            from: toHex(log.contract.getAddress()),
            to: toHex(to),
            value: '0x' + valueBigInt.toString(16),
          };
          this.transfers.unshift(transfer);
        }
      }
    } else if (toHex(to) == '0x00000000000000000000000000000000000000fd') {
      // This is the transfer precompile "address", inspect its arguments
      const stackOffset = 1;
      const inputOffset = log.stack.peek(2 + stackOffset).valueOf();
      const inputLength = log.stack.peek(3 + stackOffset).valueOf();
      const inputEnd = inputOffset + inputLength;
      const input = toHex(log.memory.slice(inputOffset, inputEnd));
      const transfer = {
        type: 'cGLD transfer precompile',
        from: '0x'+input.slice(2+24, 2+64),
        to: '0x'+input.slice(2+64+24, 2+64*2),
        value: '0x'+input.slice(2+64*2, 2+64*3),
      };
      this.transfers.push(transfer);
    }
  },

  // result is invoked when all the opcodes have been iterated over and returns
  // the final result of the tracing.
  result(ctx, db) {
    if (this.reverted) {
      this.transfers = []
    } else if (ctx.type == 'CALL') {
      valueBigInt = bigInt(ctx.value.toString());
      if (valueBigInt.gt(0)) {
        const transfer = {
          type: 'cGLD transfer',
          from: toHex(ctx.from),
          to: toHex(ctx.to),
          value: '0x' + valueBigInt.toString(16),
        };
        this.transfers.unshift(transfer);
      }
    }
    // Return in same format as callTracer: -calls, +transfers, and +block
    return {
      type:      ctx.type,
      from:      toHex(ctx.from),
      to:        toHex(ctx.to),
      value:     '0x' + ctx.value.toString(16),
      gas:       '0x' + bigInt(ctx.gas).toString(16),
      gasUsed:   '0x' + bigInt(ctx.gasUsed).toString(16),
      input:     toHex(ctx.input),
      output:    toHex(ctx.output),
      block:     ctx.block,
      time:      ctx.time,
      transfers: this.transfers,
    };
  },
}`

type transferTracerResponse struct {
	Transfers []Transfer `json:"tranfers"`
}

type Transfer struct {
	From  common.Address `json:"from"`
	To    common.Address `json:"from"`
	Value *big.Int       `json:"value"`
}

// UnmarshalJSON unmarshals from JSON.
func (t *Transfer) UnmarshalJSON(input []byte) error {
	type Transfer struct {
		From  common.Address `json:"from"`
		To    common.Address `json:"from"`
		Value *hexutil.Big   `json:"value"`
	}
	var dec Transfer
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}

	t.From = dec.From
	t.To = dec.To
	t.Value = (*big.Int)(dec.Value)
	if dec.Value == nil {
		return errors.New("missing required field 'value' for Transfer")
	}
	return nil
}

func (dc *DebugClient) TransactionTransfers(ctx context.Context, txhash common.Hash) ([]Transfer, error) {
	tracerConfig := &eth.TraceConfig{Tracer: &transferTracer}
	var response transferTracerResponse

	err := dc.TraceTransaction(ctx, &response, txhash, tracerConfig)
	if err != nil {
		return nil, err
	}
	return response.Transfers, nil
}
