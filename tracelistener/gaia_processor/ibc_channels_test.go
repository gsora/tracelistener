package gaia_processor

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	models "github.com/allinbits/demeris-backend-models/tracelistener"
	"github.com/allinbits/tracelistener/tracelistener"
	"github.com/allinbits/tracelistener/tracelistener/config"
)

func TestIbcChannelsProcess(t *testing.T) {
	i := ibcChannelsProcessor{}

	// test ownkey prefix
	require.True(t, i.OwnsKey(append([]byte(host.KeyChannelEndPrefix), []byte("key")...)))
	require.False(t, i.OwnsKey(append([]byte("0x0"), []byte("key")...)))

	DataProcessor, _ := New(zap.NewNop().Sugar(), &config.Config{})
	gp := DataProcessor.(*Processor)
	require.NotNil(t, gp)
	p.cdc = gp.cdc

	tests := []struct {
		name        string
		channel     types.Channel
		newMessage  tracelistener.TraceOperation
		expectedErr bool
		expectedLen int
	}{
		{
			"Ibc channel - no error",
			types.Channel{
				State:    4,
				Ordering: 1,
				Counterparty: types.Counterparty{
					PortId:    "some",
					ChannelId: "channelIdtest",
				},
				ConnectionHops: []string{"connectionhopID"},
			},
			tracelistener.TraceOperation{
				Operation: string(tracelistener.WriteOp),
				Key:       []byte("cosmos/ports/x/channels/ibc"),
			},
			false,
			1,
		},
		{
			"Cannot parse channel path - error",
			types.Channel{
				State:    4,
				Ordering: 1,
				Counterparty: types.Counterparty{
					PortId:    "some",
					ChannelId: "channelIdtest",
				},
				ConnectionHops: []string{"connectionhopID"},
			},
			tracelistener.TraceOperation{
				Operation: string(tracelistener.WriteOp),
				Key:       []byte("cosmos/x/channels/ibc"),
			},
			true,
			0,
		},
		{
			"Invalid counterparty port ID - error",
			types.Channel{
				State:    4,
				Ordering: 1,
				Counterparty: types.Counterparty{
					PortId:    "",
					ChannelId: "channelIdtest",
				},
				ConnectionHops: []string{"connectionhopID"},
			},
			tracelistener.TraceOperation{
				Operation: string(tracelistener.WriteOp),
				Key:       []byte("cosmos/ports/x/channels/ibc"),
			},
			true,
			0,
		},
		{
			"Invalid connection hop ID - error",
			types.Channel{
				State:    4,
				Ordering: 1,
				Counterparty: types.Counterparty{
					PortId:    "some",
					ChannelId: "channelIdtest",
				},
				ConnectionHops: []string{""},
			},
			tracelistener.TraceOperation{
				Operation: string(tracelistener.WriteOp),
				Key:       []byte("cosmos/ports/x/channels/ibc"),
			},
			true,
			0,
		},
		{
			"Invalid channel state - error",
			types.Channel{
				State:    0,
				Ordering: 1,
				Counterparty: types.Counterparty{
					PortId:    "some",
					ChannelId: "channelIdtest",
				},
				ConnectionHops: []string{"connectionhopID"},
			},
			tracelistener.TraceOperation{
				Operation: string(tracelistener.WriteOp),
				Key:       []byte("cosmos/ports/x/channels/ibc"),
			},
			true,
			0,
		},
		{
			"Invalid channel ordering - error",
			types.Channel{
				State:    4,
				Ordering: 0,
				Counterparty: types.Counterparty{
					PortId:    "some",
					ChannelId: "channelIdtest",
				},
				ConnectionHops: []string{"connectionhopID"},
			},
			tracelistener.TraceOperation{
				Operation: string(tracelistener.WriteOp),
				Key:       []byte("cosmos/ports/x/channels/ibc"),
			},
			true,
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i.channelsCache = map[channelCacheEntry]models.IBCChannelRow{}
			i.l = zap.NewNop().Sugar()

			value, err := p.cdc.MarshalBinaryBare(&tt.channel)
			require.NoError(t, err)
			tt.newMessage.Value = value

			err = i.Process(tt.newMessage)
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// check cache length
			require.Len(t, i.channelsCache, tt.expectedLen)

			// if channelcache not empty then check the data
			for k, _ := range i.channelsCache {
				row := i.channelsCache[channelCacheEntry{channelID: k.channelID, portID: k.portID}]
				require.NotNil(t, row)

				state := row.State
				require.Equal(t, int32(tt.channel.State), state)
				return
			}
		})
	}
}
