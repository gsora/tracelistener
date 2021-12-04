package gaia_processor

import (
	"bytes"
	"encoding/hex"

	models "github.com/allinbits/demeris-backend-models/tracelistener"
	"github.com/allinbits/tracelistener/tracelistener"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"go.uber.org/zap"
)

type validatorCacheEntry struct {
	operator string
}
type validatorsProcessor struct {
	l                     *zap.SugaredLogger
	insertValidatorsCache map[validatorCacheEntry]models.ValidatorRow
	deleteValidatorsCache map[validatorCacheEntry]models.ValidatorRow
}

func (*validatorsProcessor) TableSchema() string {
	return createValidatorsTable
}

func (vp *validatorsProcessor) ModuleName() string {
	return "validators"
}

func (vp *validatorsProcessor) FlushCache() []tracelistener.WritebackOp {

	if len(vp.insertValidatorsCache) == 0 && len(vp.deleteValidatorsCache) == 0 {
		return nil
	}

	insertValidators := make([]models.DatabaseEntrier, 0, len(vp.insertValidatorsCache))
	deleteValidators := make([]models.DatabaseEntrier, 0, len(vp.deleteValidatorsCache))

	if len(vp.insertValidatorsCache) != 0 {
		for _, v := range vp.insertValidatorsCache {
			insertValidators = append(insertValidators, v)
		}
	}

	vp.insertValidatorsCache = map[validatorCacheEntry]models.ValidatorRow{}

	if len(vp.deleteValidatorsCache) != 0 {
		for _, v := range vp.deleteValidatorsCache {
			deleteValidators = append(deleteValidators, v)
		}
	}

	vp.deleteValidatorsCache = map[validatorCacheEntry]models.ValidatorRow{}

	return []tracelistener.WritebackOp{
		{
			DatabaseExec: insertValidator,
			Data:         insertValidators,
		},
		{
			DatabaseExec: deleteValidator,
			Data:         deleteValidators,
		},
	}
}
func (vp *validatorsProcessor) OwnsKey(key []byte) bool {
	return bytes.HasPrefix(key, types.ValidatorsKey)
}

func (vp *validatorsProcessor) Process(data tracelistener.TraceOperation) error {

	if data.Operation == tracelistener.DeleteOp.String() {
		if len(data.Key) < 21 {
			return nil
		}

		operatorAddress := hex.EncodeToString(data.Key[1:21])
		vp.l.Debugw("new validator delete", "operator address", operatorAddress)

		vp.deleteValidatorsCache[validatorCacheEntry{
			operator: operatorAddress,
		}] = models.ValidatorRow{
			OperatorAddress: operatorAddress,
		}

		return nil
	}

	v := types.Validator{}

	if err := p.cdc.UnmarshalBinaryBare(data.Value, &v); err != nil {
		return err
	}

	val := string(v.ConsensusPubkey.GetValue())

	k := hex.EncodeToString(data.Key)

	vp.l.Debugw("new validator write",
		"operator_address", v.OperatorAddress,
		"height", data.BlockHeight,
		"txHash", data.TxHash,
		"cons pub key type", data.TxHash,
		"cons pub key", val,
		"key", k,
	)

	vp.insertValidatorsCache[validatorCacheEntry{
		operator: v.OperatorAddress,
	}] = models.ValidatorRow{
		OperatorAddress:      v.OperatorAddress,
		ConsensusPubKeyType:  v.ConsensusPubkey.GetTypeUrl(),
		ConsensusPubKeyValue: v.ConsensusPubkey.Value,
		Jailed:               v.Jailed,
		Status:               int32(v.Status),
		Tokens:               v.Tokens.String(),
		DelegatorShares:      v.DelegatorShares.String(),
		Moniker:              v.Description.Moniker,
		Identity:             v.Description.Identity,
		Website:              v.Description.Website,
		SecurityContact:      v.Description.SecurityContact,
		Details:              v.Description.Details,
		UnbondingHeight:      v.UnbondingHeight,
		UnbondingTime:        v.UnbondingTime.String(),
		CommissionRate:       v.Commission.CommissionRates.Rate.String(),
		MaxRate:              v.Commission.CommissionRates.MaxRate.String(),
		MaxChangeRate:        v.Commission.CommissionRates.MaxChangeRate.String(),
		UpdateTime:           v.Commission.UpdateTime.String(),
		MinSelfDelegation:    v.MinSelfDelegation.String(),
	}

	return nil
}
