package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	merlion "github.com/merlion-zone/merlion/types"
	"github.com/merlion-zone/merlion/x/maker/types"
)

func (suite *KeeperTestSuite) TestEstimateMintBySwapIn() {
	testCases := []struct {
		name     string
		malleate func()
		req      *types.EstimateMintBySwapInRequest
		expPass  bool
		expErr   error
		expRes   *types.EstimateMintBySwapInResponse
	}{
		{
			name: "mer price too low",
			malleate: func() {
				suite.app.OracleKeeper.SetExchangeRate(suite.ctx, merlion.MicroUSMDenom, sdk.NewDecWithPrec(989, 3))
			},
			req:     &types.EstimateMintBySwapInRequest{BackingDenom: suite.bcDenom},
			expPass: false,
			expErr:  types.ErrMerPriceTooLow,
		},
		{
			name:    "backing denom not found",
			req:     &types.EstimateMintBySwapInRequest{BackingDenom: "fil"},
			expPass: false,
			expErr:  types.ErrBackingCoinNotFound,
		},
		{
			name:    "backing denom disabled",
			req:     &types.EstimateMintBySwapInRequest{BackingDenom: "eth"},
			expPass: false,
			expErr:  types.ErrBackingCoinDisabled,
		},
		{
			name: "mer over ceiling",
			req: &types.EstimateMintBySwapInRequest{
				MintOut:      sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(2_500000)),
				BackingDenom: suite.bcDenom,
			},
			expPass: false,
			expErr:  types.ErrMerCeiling,
		},
		{
			name: "default full backing",
			req: &types.EstimateMintBySwapInRequest{
				MintOut:      sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(1_000000)),
				BackingDenom: suite.bcDenom,
			},
			expPass: true,
			expRes: &types.EstimateMintBySwapInResponse{
				BackingIn: sdk.NewCoin(suite.bcDenom, sdk.NewInt(1015152)), // 1_000000 * (1+0.005) / 0.99
				LionIn:    sdk.NewCoin(merlion.AttoLionDenom, sdk.ZeroInt()),
				MintFee:   sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(5000)),
			},
		},
		{
			name: "user asked full backing",
			malleate: func() {
				suite.app.MakerKeeper.SetBackingRatio(suite.ctx, sdk.NewDecWithPrec(80, 2))
			},
			req: &types.EstimateMintBySwapInRequest{
				MintOut:      sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(1_000000)),
				BackingDenom: suite.bcDenom,
				FullBacking:  true,
			},
			expPass: true,
			expRes: &types.EstimateMintBySwapInResponse{
				BackingIn: sdk.NewCoin(suite.bcDenom, sdk.NewInt(1015152)), // 1_000000 * (1+0.005) / 0.99
				LionIn:    sdk.NewCoin(merlion.AttoLionDenom, sdk.ZeroInt()),
				MintFee:   sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(5000)),
			},
		},
		{
			name: "full algorithmic",
			malleate: func() {
				suite.app.MakerKeeper.SetBackingRatio(suite.ctx, sdk.ZeroDec())
			},
			req: &types.EstimateMintBySwapInRequest{
				MintOut:      sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(1_000000)),
				BackingDenom: suite.bcDenom,
			},
			expPass: true,
			expRes: &types.EstimateMintBySwapInResponse{
				BackingIn: sdk.NewCoin(suite.bcDenom, sdk.ZeroInt()),
				LionIn:    sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(10050_000000_000000)), // 1_000000 * (1+0.005) / 10**-10
				MintFee:   sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(5000)),
			},
		},
		{
			name: "fractional",
			malleate: func() {
				suite.app.MakerKeeper.SetBackingRatio(suite.ctx, sdk.NewDecWithPrec(80, 2))
			},
			req: &types.EstimateMintBySwapInRequest{
				MintOut:      sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(1_000000)),
				BackingDenom: suite.bcDenom,
			},
			expPass: true,
			expRes: &types.EstimateMintBySwapInResponse{
				BackingIn: sdk.NewCoin(suite.bcDenom, sdk.NewInt(812121)),                     // 1_000000 * (1+0.005) * 0.8 / 0.99
				LionIn:    sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(2010_000000_000000)), // 1_000000 * (1+0.005) * 0.2 / 10**-10
				MintFee:   sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(5000)),
			},
		},
		{
			name: "backing over ceiling",
			req: &types.EstimateMintBySwapInRequest{
				MintOut:      sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(1_500000)),
				BackingDenom: suite.bcDenom,
			},
			expPass: false,
			expErr:  types.ErrBackingCeiling,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			suite.setupEstimationTest()
			if tc.malleate != nil {
				tc.malleate()
			}

			ctx := sdk.WrapSDKContext(suite.ctx)
			res, err := suite.queryClient.EstimateMintBySwapIn(ctx, tc.req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expRes, res)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.expErr)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestEstimateMintBySwapOut() {
	testCases := []struct {
		name     string
		malleate func()
		req      *types.EstimateMintBySwapOutRequest
		expPass  bool
		expErr   error
		expRes   *types.EstimateMintBySwapOutResponse
	}{
		{
			name: "mer price too low",
			malleate: func() {
				suite.app.OracleKeeper.SetExchangeRate(suite.ctx, merlion.MicroUSMDenom, sdk.NewDecWithPrec(989, 3))
			},
			req:     &types.EstimateMintBySwapOutRequest{BackingInMax: sdk.NewCoin(suite.bcDenom, sdk.ZeroInt())},
			expPass: false,
			expErr:  types.ErrMerPriceTooLow,
		},
		{
			name:    "backing denom not found",
			req:     &types.EstimateMintBySwapOutRequest{BackingInMax: sdk.NewCoin("fil", sdk.ZeroInt())},
			expPass: false,
			expErr:  types.ErrBackingCoinNotFound,
		},
		{
			name:    "backing denom disabled",
			req:     &types.EstimateMintBySwapOutRequest{BackingInMax: sdk.NewCoin("eth", sdk.ZeroInt())},
			expPass: false,
			expErr:  types.ErrBackingCoinDisabled,
		},
		{
			name: "default full backing",
			req: &types.EstimateMintBySwapOutRequest{
				BackingInMax: sdk.NewCoin(suite.bcDenom, sdk.NewInt(1_000000)),
			},
			expPass: true,
			expRes: &types.EstimateMintBySwapOutResponse{
				BackingIn: sdk.NewCoin(suite.bcDenom, sdk.NewInt(1_000000)),
				LionIn:    sdk.NewCoin(merlion.AttoLionDenom, sdk.ZeroInt()),
				MintOut:   sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(985075)), // 1_000000 * 0.99 * (1 / (1+0.005))
				MintFee:   sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(4925)),   // 1_000000 * 0.99 * (0.005 / (1+0.005))
			},
		},
		{
			name: "user asked full backing",
			malleate: func() {
				suite.app.MakerKeeper.SetBackingRatio(suite.ctx, sdk.NewDecWithPrec(80, 2))
			},
			req: &types.EstimateMintBySwapOutRequest{
				BackingInMax: sdk.NewCoin(suite.bcDenom, sdk.NewInt(1_000000)),
				FullBacking:  true,
			},
			expPass: true,
			expRes: &types.EstimateMintBySwapOutResponse{
				BackingIn: sdk.NewCoin(suite.bcDenom, sdk.NewInt(1_000000)),
				LionIn:    sdk.NewCoin(merlion.AttoLionDenom, sdk.ZeroInt()),
				MintOut:   sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(985075)), // 1_000000 * 0.99 * (1 / (1+0.005))
				MintFee:   sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(4925)),   // 1_000000 * 0.99 * (0.005 / (1+0.005))
			},
		},
		{
			name: "full algorithmic",
			malleate: func() {
				suite.app.MakerKeeper.SetBackingRatio(suite.ctx, sdk.ZeroDec())
			},
			req: &types.EstimateMintBySwapOutRequest{
				BackingInMax: sdk.NewCoin(suite.bcDenom, sdk.NewInt(1_000000)),
				LionInMax:    sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(10000_000000_000000)),
			},
			expPass: true,
			expRes: &types.EstimateMintBySwapOutResponse{
				BackingIn: sdk.NewCoin(suite.bcDenom, sdk.ZeroInt()),
				LionIn:    sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(10000_000000_000000)),
				MintOut:   sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(995025)), // 10**16 * 10**-10 * (1 / (1+0.005))
				MintFee:   sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(4975)),   // 10**16 * 10**-10 * (0.005 / (1+0.005))
			},
		},
		{
			name: "zero lion using backing",
			malleate: func() {
				suite.app.MakerKeeper.SetBackingRatio(suite.ctx, sdk.NewDecWithPrec(80, 2))
			},
			req: &types.EstimateMintBySwapOutRequest{
				BackingInMax: sdk.NewCoin(suite.bcDenom, sdk.NewInt(500000)),
				LionInMax:    sdk.NewCoin(merlion.AttoLionDenom, sdk.ZeroInt()),
			},
			expPass: true,
			expRes: &types.EstimateMintBySwapOutResponse{
				BackingIn: sdk.NewCoin(suite.bcDenom, sdk.NewInt(500000)),
				LionIn:    sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(1237_500000_000000)), // 500000 * 0.99 / 0.8 * 0.2 / (10**-10)
				MintOut:   sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(615672)),             // 500000 * 0.99 / 0.8 * (1 / (1+0.005))
				MintFee:   sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(3078)),               // 500000 * 0.99 / 0.8 * (0.005 / (1+0.005))
			},
		},
		{
			name: "fractional using max backing",
			malleate: func() {
				suite.app.MakerKeeper.SetBackingRatio(suite.ctx, sdk.NewDecWithPrec(80, 2))
			},
			req: &types.EstimateMintBySwapOutRequest{
				BackingInMax: sdk.NewCoin(suite.bcDenom, sdk.NewInt(500000)),
				LionInMax:    sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(10000_000000_000000)),
			},
			expPass: true,
			expRes: &types.EstimateMintBySwapOutResponse{
				BackingIn: sdk.NewCoin(suite.bcDenom, sdk.NewInt(500000)),
				LionIn:    sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(1237_500000_000000)), // 500000 * 0.99 / 0.8 * 0.2 / (10**-10)
				MintOut:   sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(615672)),             // 500000 * 0.99 / 0.8 * (1 / (1+0.005))
				MintFee:   sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(3078)),               // 500000 * 0.99 / 0.8 * (0.005 / (1+0.005))
			},
		},
		{
			name: "zero backing using lion",
			malleate: func() {
				suite.app.MakerKeeper.SetBackingRatio(suite.ctx, sdk.NewDecWithPrec(20, 2))
			},
			req: &types.EstimateMintBySwapOutRequest{
				BackingInMax: sdk.NewCoin(suite.bcDenom, sdk.ZeroInt()),
				LionInMax:    sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(10000_000000_000000)),
			},
			expPass: true,
			expRes: &types.EstimateMintBySwapOutResponse{
				BackingIn: sdk.NewCoin(suite.bcDenom, sdk.NewInt(252525)), // 10**16 * 10**-10 / 0.8 * 0.2 / 0.99
				LionIn:    sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(10000_000000_000000)),
				MintOut:   sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(1243781)), // 10**16 * 10**-10 / 0.8 * (1 / (1+0.005))
				MintFee:   sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(6219)),    // 10**16 * 10**-10 / 0.8 * (0.005 / (1+0.005))
			},
		},
		{
			name: "fractional using max lion",
			malleate: func() {
				suite.app.MakerKeeper.SetBackingRatio(suite.ctx, sdk.NewDecWithPrec(20, 2))
			},
			req: &types.EstimateMintBySwapOutRequest{
				BackingInMax: sdk.NewCoin(suite.bcDenom, sdk.NewInt(1_000000)),
				LionInMax:    sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(10000_000000_000000)),
			},
			expPass: true,
			expRes: &types.EstimateMintBySwapOutResponse{
				BackingIn: sdk.NewCoin(suite.bcDenom, sdk.NewInt(252525)), // 10**16 * 10**-10 / 0.8 * 0.2 / 0.99
				LionIn:    sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(10000_000000_000000)),
				MintOut:   sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(1243781)), // 10**16 * 10**-10 / 0.8 * (1 / (1+0.005))
				MintFee:   sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(6219)),    // 10**16 * 10**-10 / 0.8 * (0.005 / (1+0.005))
			},
		},
		{
			name: "mer over ceiling",
			req: &types.EstimateMintBySwapOutRequest{
				BackingInMax: sdk.NewCoin(suite.bcDenom, sdk.NewInt(2_500000)),
			},
			expPass: false,
			expErr:  types.ErrMerCeiling,
		},
		{
			name: "backing over ceiling",
			req: &types.EstimateMintBySwapOutRequest{
				BackingInMax: sdk.NewCoin(suite.bcDenom, sdk.NewInt(1_500000)),
			},
			expPass: false,
			expErr:  types.ErrBackingCeiling,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			suite.setupEstimationTest()
			if tc.malleate != nil {
				tc.malleate()
			}

			ctx := sdk.WrapSDKContext(suite.ctx)
			res, err := suite.queryClient.EstimateMintBySwapOut(ctx, tc.req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expRes, res)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.expErr)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestEstimateBurnBySwapIn() {
	testCases := []struct {
		name     string
		malleate func()
		req      *types.EstimateBurnBySwapInRequest
		expPass  bool
		expErr   error
		expRes   *types.EstimateBurnBySwapInResponse
	}{
		{
			name: "mer price too high",
			malleate: func() {
				suite.app.OracleKeeper.SetExchangeRate(suite.ctx, merlion.MicroUSMDenom, sdk.NewDecWithPrec(1011, 3))
			},
			req:     &types.EstimateBurnBySwapInRequest{BackingOutMax: sdk.NewCoin(suite.bcDenom, sdk.ZeroInt())},
			expPass: false,
			expErr:  types.ErrMerPriceTooHigh,
		},
		{
			name:    "backing denom not found",
			req:     &types.EstimateBurnBySwapInRequest{BackingOutMax: sdk.NewCoin("fil", sdk.ZeroInt())},
			expPass: false,
			expErr:  types.ErrBackingCoinNotFound,
		},
		{
			name:    "backing denom disabled",
			req:     &types.EstimateBurnBySwapInRequest{BackingOutMax: sdk.NewCoin("eth", sdk.ZeroInt())},
			expPass: false,
			expErr:  types.ErrBackingCoinDisabled,
		},
		{
			name:    "moudle backing insufficient",
			req:     &types.EstimateBurnBySwapInRequest{BackingOutMax: sdk.NewCoin(suite.bcDenom, sdk.NewInt(1_000000))},
			expPass: false,
			expErr:  types.ErrBackingCoinInsufficient,
		},
		{
			name: "full backing",
			malleate: func() {
				suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(suite.bcDenom, sdk.NewInt(1000_000000))))
			},
			req: &types.EstimateBurnBySwapInRequest{
				BackingOutMax: sdk.NewCoin(suite.bcDenom, sdk.NewInt(1_000000)),
			},
			expPass: true,
			expRes: &types.EstimateBurnBySwapInResponse{
				BurnIn:     sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(995976)), // 1_000000 * 0.99 / (1-0.006)
				BackingOut: sdk.NewCoin(suite.bcDenom, sdk.NewInt(1_000000)),
				LionOut:    sdk.NewCoin(merlion.AttoLionDenom, sdk.ZeroInt()),
				BurnFee:    sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(5976)), // 1_000000 * 0.99 / (1-0.006) * 0.006
			},
		},
		{
			name: "full algorithmic",
			malleate: func() {
				suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(suite.bcDenom, sdk.NewInt(1000_000000))))
				suite.app.MakerKeeper.SetBackingRatio(suite.ctx, sdk.ZeroDec())
			},
			req: &types.EstimateBurnBySwapInRequest{
				BackingOutMax: sdk.NewCoin(suite.bcDenom, sdk.NewInt(1_000000)),
				LionOutMax:    sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(10000_000000_000000)),
			},
			expPass: true,
			expRes: &types.EstimateBurnBySwapInResponse{
				BurnIn:     sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(1006036)), // 10**16 * 10**-10 / (1-0.006)
				BackingOut: sdk.NewCoin(suite.bcDenom, sdk.ZeroInt()),
				LionOut:    sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(10000_000000_000000)),
				BurnFee:    sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(6036)), // 10**16 * 10**-10 / (1-0.006) * 0.006
			},
		},
		{
			name: "zero lion using backing",
			malleate: func() {
				suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(suite.bcDenom, sdk.NewInt(1000_000000))))
				suite.app.MakerKeeper.SetBackingRatio(suite.ctx, sdk.NewDecWithPrec(80, 2))
			},
			req: &types.EstimateBurnBySwapInRequest{
				BackingOutMax: sdk.NewCoin(suite.bcDenom, sdk.NewInt(500000)),
				LionOutMax:    sdk.NewCoin(merlion.AttoLionDenom, sdk.ZeroInt()),
			},
			expPass: true,
			expRes: &types.EstimateBurnBySwapInResponse{
				BurnIn:     sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(622485)), // 500000 * 0.99 / 0.8 / (1-0.006)
				BackingOut: sdk.NewCoin(suite.bcDenom, sdk.NewInt(500000)),
				LionOut:    sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(1237_500000_000000)), // 500000 * 0.99 / 0.8 * 0.2 / (10**-10)
				BurnFee:    sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(3735)),               // 500000 * 0.99 / 0.8 / (1-0.006) * 0.006
			},
		},
		{
			name: "fractional using max backing",
			malleate: func() {
				suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(suite.bcDenom, sdk.NewInt(1000_000000))))
				suite.app.MakerKeeper.SetBackingRatio(suite.ctx, sdk.NewDecWithPrec(80, 2))
			},
			req: &types.EstimateBurnBySwapInRequest{
				BackingOutMax: sdk.NewCoin(suite.bcDenom, sdk.NewInt(500000)),
				LionOutMax:    sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(10000_000000_000000)),
			},
			expPass: true,
			expRes: &types.EstimateBurnBySwapInResponse{
				BurnIn:     sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(622485)), // 500000 * 0.99 / 0.8 / (1-0.006)
				BackingOut: sdk.NewCoin(suite.bcDenom, sdk.NewInt(500000)),
				LionOut:    sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(1237_500000_000000)), // 500000 * 0.99 / 0.8 * 0.2 / (10**-10)
				BurnFee:    sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(3735)),               // 500000 * 0.99 / 0.8 / (1-0.006) * 0.006
			},
		},
		{
			name: "zero backing using lion",
			malleate: func() {
				suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(suite.bcDenom, sdk.NewInt(1000_000000))))
				suite.app.MakerKeeper.SetBackingRatio(suite.ctx, sdk.NewDecWithPrec(20, 2))
			},
			req: &types.EstimateBurnBySwapInRequest{
				BackingOutMax: sdk.NewCoin(suite.bcDenom, sdk.ZeroInt()),
				LionOutMax:    sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(10000_000000_000000)),
			},
			expPass: true,
			expRes: &types.EstimateBurnBySwapInResponse{
				BurnIn:     sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(1257545)), // 10**16 * 10**-10 / 0.8 / (1-0.006)
				BackingOut: sdk.NewCoin(suite.bcDenom, sdk.NewInt(252525)),          // 10**16 * 10**-10 / 0.8 * 0.2 / 0.99
				LionOut:    sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(10000_000000_000000)),
				BurnFee:    sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(7545)), // 10**16 * 10**-10 / 0.8 / (1-0.006) * 0.006
			},
		},
		{
			name: "fractional using max lion",
			malleate: func() {
				suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(suite.bcDenom, sdk.NewInt(1000_000000))))
				suite.app.MakerKeeper.SetBackingRatio(suite.ctx, sdk.NewDecWithPrec(20, 2))
			},
			req: &types.EstimateBurnBySwapInRequest{
				BackingOutMax: sdk.NewCoin(suite.bcDenom, sdk.NewInt(1_000000)),
				LionOutMax:    sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(10000_000000_000000)),
			},
			expPass: true,
			expRes: &types.EstimateBurnBySwapInResponse{
				BurnIn:     sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(1257545)), // 10**16 * 10**-10 / 0.8 / (1-0.006)
				BackingOut: sdk.NewCoin(suite.bcDenom, sdk.NewInt(252525)),          // 10**16 * 10**-10 / 0.8 * 0.2 / 0.99
				LionOut:    sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(10000_000000_000000)),
				BurnFee:    sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(7545)), // 10**16 * 10**-10 / 0.8 / (1-0.006) * 0.006
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			suite.setupEstimationTest()
			if tc.malleate != nil {
				tc.malleate()
			}

			ctx := sdk.WrapSDKContext(suite.ctx)
			res, err := suite.queryClient.EstimateBurnBySwapIn(ctx, tc.req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expRes, res)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.expErr)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestEstimateBurnBySwapOut() {
	testCases := []struct {
		name     string
		malleate func()
		req      *types.EstimateBurnBySwapOutRequest
		expPass  bool
		expErr   error
		expRes   *types.EstimateBurnBySwapOutResponse
	}{
		{
			name: "mer price too high",
			malleate: func() {
				suite.app.OracleKeeper.SetExchangeRate(suite.ctx, merlion.MicroUSMDenom, sdk.NewDecWithPrec(1011, 3))
			},
			req:     &types.EstimateBurnBySwapOutRequest{BackingDenom: suite.bcDenom},
			expPass: false,
			expErr:  types.ErrMerPriceTooHigh,
		},
		{
			name:    "backing denom not found",
			req:     &types.EstimateBurnBySwapOutRequest{BackingDenom: "fil"},
			expPass: false,
			expErr:  types.ErrBackingCoinNotFound,
		},
		{
			name:    "backing denom disabled",
			req:     &types.EstimateBurnBySwapOutRequest{BackingDenom: "eth"},
			expPass: false,
			expErr:  types.ErrBackingCoinDisabled,
		},
		{
			name: "moudle backing insufficient",
			req: &types.EstimateBurnBySwapOutRequest{
				BurnIn:       sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(1)),
				BackingDenom: suite.bcDenom,
			},
			expPass: false,
			expErr:  types.ErrBackingCoinInsufficient,
		},
		{
			name: "full backing",
			malleate: func() {
				suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(suite.bcDenom, sdk.NewInt(1000_000000))))
			},
			req: &types.EstimateBurnBySwapOutRequest{
				BurnIn:       sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(1_000000)),
				BackingDenom: suite.bcDenom,
			},
			expPass: true,
			expRes: &types.EstimateBurnBySwapOutResponse{
				BackingOut: sdk.NewCoin(suite.bcDenom, sdk.NewInt(1_004040)), // 1_000000 * (1-0.006) / 0.99
				LionOut:    sdk.NewCoin(merlion.AttoLionDenom, sdk.ZeroInt()),
				BurnFee:    sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(6000)), // 1_000000 * 0.006
			},
		},
		{
			name: "full algorithmic",
			malleate: func() {
				suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(suite.bcDenom, sdk.NewInt(1000_000000))))
				suite.app.MakerKeeper.SetBackingRatio(suite.ctx, sdk.ZeroDec())
			},
			req: &types.EstimateBurnBySwapOutRequest{
				BurnIn:       sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(1_000000)),
				BackingDenom: suite.bcDenom,
			},
			expPass: true,
			expRes: &types.EstimateBurnBySwapOutResponse{
				BackingOut: sdk.NewCoin(suite.bcDenom, sdk.ZeroInt()),                          // 1_000000 * (1-0.006) / 0.99
				LionOut:    sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(9940_000000_000000)), // 1_000000 * (1-0.006) / 10**-10
				BurnFee:    sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(6000)),               // 1_000000 * 0.006
			},
		},
		{
			name: "fractional",
			malleate: func() {
				suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(suite.bcDenom, sdk.NewInt(1000_000000))))
				suite.app.MakerKeeper.SetBackingRatio(suite.ctx, sdk.NewDecWithPrec(80, 2))
			},
			req: &types.EstimateBurnBySwapOutRequest{
				BurnIn:       sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(1_000000)),
				BackingDenom: suite.bcDenom,
			},
			expPass: true,
			expRes: &types.EstimateBurnBySwapOutResponse{
				BackingOut: sdk.NewCoin(suite.bcDenom, sdk.NewInt(803232)),                     // 1_000000 * (1-0.006) * 0.8 / 0.99
				LionOut:    sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(19880_00000_000000)), // 1_000000 * (1-0.006) * 0.2 / 10**-10
				BurnFee:    sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(6000)),               // 1_000000 * 0.006
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			suite.setupEstimationTest()
			if tc.malleate != nil {
				tc.malleate()
			}

			ctx := sdk.WrapSDKContext(suite.ctx)
			res, err := suite.queryClient.EstimateBurnBySwapOut(ctx, tc.req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expRes, res)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.expErr)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestEstimateBuyBackingIn() {
	testCases := []struct {
		name     string
		malleate func()
		req      *types.EstimateBuyBackingInRequest
		expPass  bool
		expErr   error
		expRes   *types.EstimateBuyBackingInResponse
	}{
		{
			name:    "backing denom not found",
			req:     &types.EstimateBuyBackingInRequest{BackingOut: sdk.NewCoin("fil", sdk.ZeroInt())},
			expPass: false,
			expErr:  types.ErrBackingCoinNotFound,
		},
		{
			name:    "backing denom disabled",
			req:     &types.EstimateBuyBackingInRequest{BackingOut: sdk.NewCoin("eth", sdk.ZeroInt())},
			expPass: false,
			expErr:  types.ErrBackingCoinDisabled,
		},
		{
			name: "no mer minted",
			malleate: func() {
				totalBacking, found := suite.app.MakerKeeper.GetTotalBacking(suite.ctx)
				suite.Require().True(found)
				totalBacking.MerMinted.Amount = sdk.ZeroInt()
				suite.app.MakerKeeper.SetTotalBacking(suite.ctx, totalBacking)
			},
			req:     &types.EstimateBuyBackingInRequest{BackingOut: sdk.NewCoin(suite.bcDenom, sdk.ZeroInt())},
			expPass: false,
			expErr:  types.ErrBackingCoinInsufficient,
		},
		{
			name: "excess backing insufficient",
			req: &types.EstimateBuyBackingInRequest{
				BackingOut: sdk.NewCoin(suite.bcDenom, sdk.NewInt(500000)),
			},
			expPass: false, // 5*10**5 * 0.99 / (1-0.007) > 9*10**6 * 0.99 * 1 - 8.5*10**6
			expErr:  types.ErrBackingCoinInsufficient,
		},
		{
			name: "pool backing insufficient",
			malleate: func() {
				poolBacking, found := suite.app.MakerKeeper.GetPoolBacking(suite.ctx, suite.bcDenom)
				suite.Require().True(found)
				poolBacking.Backing.Amount = sdk.ZeroInt()
				suite.app.MakerKeeper.SetPoolBacking(suite.ctx, poolBacking)
			},
			req: &types.EstimateBuyBackingInRequest{
				BackingOut: sdk.NewCoin(suite.bcDenom, sdk.NewInt(300000)),
			},
			expPass: false,
			expErr:  types.ErrBackingCoinInsufficient,
		},
		{
			name: "correct",
			req: &types.EstimateBuyBackingInRequest{
				BackingOut: sdk.NewCoin(suite.bcDenom, sdk.NewInt(300000)),
			},
			expPass: true,
			expRes: &types.EstimateBuyBackingInResponse{
				LionIn:     sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(2990_936555_891239)), // 3*10**5 * 0.99 / (1-0.007) / 10**-10
				BuybackFee: sdk.NewCoin(suite.bcDenom, sdk.NewInt(2094)),                       // 3*10**5 * 0.99 / (1-0.007) * 0.007
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			suite.setupEstimationTest()
			if tc.malleate != nil {
				tc.malleate()
			}

			ctx := sdk.WrapSDKContext(suite.ctx)
			res, err := suite.queryClient.EstimateBuyBackingIn(ctx, tc.req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expRes, res)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.expErr)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestEstimateBuyBackingOut() {
	testCases := []struct {
		name     string
		malleate func()
		req      *types.EstimateBuyBackingOutRequest
		expPass  bool
		expErr   error
		expRes   *types.EstimateBuyBackingOutResponse
	}{
		{
			name:    "backing denom not found",
			req:     &types.EstimateBuyBackingOutRequest{BackingDenom: "fil"},
			expPass: false,
			expErr:  types.ErrBackingCoinNotFound,
		},
		{
			name:    "backing denom disabled",
			req:     &types.EstimateBuyBackingOutRequest{BackingDenom: "eth"},
			expPass: false,
			expErr:  types.ErrBackingCoinDisabled,
		},
		{
			name: "no mer minted",
			malleate: func() {
				totalBacking, found := suite.app.MakerKeeper.GetTotalBacking(suite.ctx)
				suite.Require().True(found)
				totalBacking.MerMinted.Amount = sdk.ZeroInt()
				suite.app.MakerKeeper.SetTotalBacking(suite.ctx, totalBacking)
			},
			req:     &types.EstimateBuyBackingOutRequest{BackingDenom: suite.bcDenom},
			expPass: false,
			expErr:  types.ErrBackingCoinInsufficient,
		},
		{
			name: "excess backing insufficient",
			req: &types.EstimateBuyBackingOutRequest{
				BackingDenom: suite.bcDenom,
				LionIn:       sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(5000_000000_000000)),
			},
			expPass: false, // 0.5*10**16 * 10**-10 > 9*10**6 * 0.99 * 1 - 8.5*10**6
			expErr:  types.ErrBackingCoinInsufficient,
		},
		{
			name: "pool backing insufficient",
			malleate: func() {
				poolBacking, found := suite.app.MakerKeeper.GetPoolBacking(suite.ctx, suite.bcDenom)
				suite.Require().True(found)
				poolBacking.Backing.Amount = sdk.ZeroInt()
				suite.app.MakerKeeper.SetPoolBacking(suite.ctx, poolBacking)
			},
			req: &types.EstimateBuyBackingOutRequest{
				BackingDenom: suite.bcDenom,
				LionIn:       sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(3000_000000_000000)),
			},
			expPass: false,
			expErr:  types.ErrBackingCoinInsufficient,
		},
		{
			name: "correct",
			req: &types.EstimateBuyBackingOutRequest{
				BackingDenom: suite.bcDenom,
				LionIn:       sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(3000_000000_000000)),
			},
			expPass: true,
			expRes: &types.EstimateBuyBackingOutResponse{
				BackingOut: sdk.NewCoin(suite.bcDenom, sdk.NewInt(300909)), // 0.3*10**16 * 10**-10 / 0.99  * (1-0.007)
				BuybackFee: sdk.NewCoin(suite.bcDenom, sdk.NewInt(2121)),   // 0.3*10**16 * 10**-10 / 0.99  * 0.007
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			suite.setupEstimationTest()
			if tc.malleate != nil {
				tc.malleate()
			}

			ctx := sdk.WrapSDKContext(suite.ctx)
			res, err := suite.queryClient.EstimateBuyBackingOut(ctx, tc.req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expRes, res)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.expErr)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestEstimateSellBackingIn() {
	testCases := []struct {
		name     string
		malleate func()
		req      *types.EstimateSellBackingInRequest
		expPass  bool
		expErr   error
		expRes   *types.EstimateSellBackingInResponse
	}{
		{
			name:    "backing denom not found",
			req:     &types.EstimateSellBackingInRequest{BackingDenom: "fil"},
			expPass: false,
			expErr:  types.ErrBackingCoinNotFound,
		},
		{
			name:    "backing denom disabled",
			req:     &types.EstimateSellBackingInRequest{BackingDenom: "eth"},
			expPass: false,
			expErr:  types.ErrBackingCoinDisabled,
		},
		{
			name: "pool backing over ceiling",
			req: &types.EstimateSellBackingInRequest{
				LionOut:      sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(20000_000000_000000)),
				BackingDenom: suite.bcDenom,
			},
			expPass: false,
			expErr:  types.ErrBackingCeiling,
		},
		{
			name: "lion insufficient",
			req: &types.EstimateSellBackingInRequest{
				LionOut:      sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(10000_000000_000000)),
				BackingDenom: suite.bcDenom,
			},
			expPass: false,
			expErr:  types.ErrLionCoinInsufficient,
		},
		{
			name: "correct",
			malleate: func() {
				totalBacking, found := suite.app.MakerKeeper.GetTotalBacking(suite.ctx)
				suite.Require().True(found)
				totalBacking.MerMinted.Amount = sdk.NewInt(10_000000)
				suite.app.MakerKeeper.SetTotalBacking(suite.ctx, totalBacking)
			},
			req: &types.EstimateSellBackingInRequest{
				LionOut:      sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(10000_000000_000000)),
				BackingDenom: suite.bcDenom,
			},
			expPass: true,
			expRes: &types.EstimateSellBackingInResponse{
				BackingIn:   sdk.NewCoin(suite.bcDenom, sdk.NewInt(1_006578)),                 // 1*10**16 / (1+0.0075-0.004) * 10**-10 / 0.99
				SellbackFee: sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(39_860488_290982)), // 1*10**16 / (1+0.0075-0.004) * 0.004
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			suite.setupEstimationTest()
			if tc.malleate != nil {
				tc.malleate()
			}

			ctx := sdk.WrapSDKContext(suite.ctx)
			res, err := suite.queryClient.EstimateSellBackingIn(ctx, tc.req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expRes, res)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.expErr)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestEstimateSellBackingOut() {
	testCases := []struct {
		name     string
		malleate func()
		req      *types.EstimateSellBackingOutRequest
		expPass  bool
		expErr   error
		expRes   *types.EstimateSellBackingOutResponse
	}{
		{
			name:    "backing denom not found",
			req:     &types.EstimateSellBackingOutRequest{BackingIn: sdk.NewCoin("fil", sdk.ZeroInt())},
			expPass: false,
			expErr:  types.ErrBackingCoinNotFound,
		},
		{
			name:    "backing denom disabled",
			req:     &types.EstimateSellBackingOutRequest{BackingIn: sdk.NewCoin("eth", sdk.ZeroInt())},
			expPass: false,
			expErr:  types.ErrBackingCoinDisabled,
		},
		{
			name:    "pool backing over ceiling",
			req:     &types.EstimateSellBackingOutRequest{BackingIn: sdk.NewCoin(suite.bcDenom, sdk.NewInt(2_000000))},
			expPass: false,
			expErr:  types.ErrBackingCeiling,
		},
		{
			name:    "lion insufficient",
			req:     &types.EstimateSellBackingOutRequest{BackingIn: sdk.NewCoin(suite.bcDenom, sdk.NewInt(1_000000))},
			expPass: false,
			expErr:  types.ErrLionCoinInsufficient,
		},
		{
			name: "correct",
			malleate: func() {
				totalBacking, found := suite.app.MakerKeeper.GetTotalBacking(suite.ctx)
				suite.Require().True(found)
				totalBacking.MerMinted.Amount = sdk.NewInt(10_000000)
				suite.app.MakerKeeper.SetTotalBacking(suite.ctx, totalBacking)
			},
			req:     &types.EstimateSellBackingOutRequest{BackingIn: sdk.NewCoin(suite.bcDenom, sdk.NewInt(1_000000))},
			expPass: true,
			expRes: &types.EstimateSellBackingOutResponse{
				LionOut:     sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(9934_650000_000000)), // 1*10**6 * 0.99 / 10**-10 * (1+0.0075-0.004)
				SellbackFee: sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(39_600000_000000)),   // 1*10**6 * 0.99 / 10**-10 * 0.004
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			suite.setupEstimationTest()
			if tc.malleate != nil {
				tc.malleate()
			}

			ctx := sdk.WrapSDKContext(suite.ctx)
			res, err := suite.queryClient.EstimateSellBackingOut(ctx, tc.req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expRes, res)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.expErr)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestEstimateMintByCollateralIn() {
	testCases := []struct {
		name     string
		malleate func()
		req      *types.EstimateMintByCollateralInRequest
		expPass  bool
		expErr   error
		expRes   *types.EstimateMintByCollateralInResponse
	}{
		// {
		//	name: "mer price too low",
		//	malleate: func() {
		//		suite.app.OracleKeeper.SetExchangeRate(suite.ctx, merlion.MicroUSMDenom, sdk.NewDecWithPrec(989, 3))
		//	},
		//	req: &types.EstimateMintByCollateralInRequest{
		//		CollateralDenom: suite.bcDenom,
		//	},
		//	expPass: false,
		//	expErr:  types.ErrMerPriceTooLow,
		// },
		{
			name: "collateral denom not found",
			req: &types.EstimateMintByCollateralInRequest{
				CollateralDenom: "fil",
			},
			expPass: false,
			expErr:  types.ErrCollateralCoinNotFound,
		},
		{
			name: "collateral denom disabled",
			req: &types.EstimateMintByCollateralInRequest{
				CollateralDenom: "eth",
			},
			expPass: false,
			expErr:  types.ErrCollateralCoinDisabled,
		},
		{
			name: "ltv too low",
			req: &types.EstimateMintByCollateralInRequest{
				CollateralDenom: suite.bcDenom,
				Ltv:             sdk.NewDecWithPrec(4, 1),
			},
			expPass: false,
			expErr:  types.ErrLTVOutOfRange,
		},
		{
			name: "ltv too high",
			req: &types.EstimateMintByCollateralInRequest{
				CollateralDenom: suite.bcDenom,
				Ltv:             sdk.NewDecWithPrec(9, 1),
			},
			expPass: false,
			expErr:  types.ErrLTVOutOfRange,
		},
		{
			name: "mer over ceiling",
			req: &types.EstimateMintByCollateralInRequest{
				CollateralDenom: suite.bcDenom,
				Ltv:             sdk.NewDecWithPrec(5, 1),
				MintOut:         sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(2e6)),
			},
			expPass: false, // max mint total < 8e6 - 6e6 = 2e6
			expErr:  types.ErrMerCeiling,
		},
		{
			name: "correct zero collateral and lion in",
			req: &types.EstimateMintByCollateralInRequest{
				CollateralDenom: suite.bcDenom,
				Ltv:             sdk.NewDecWithPrec(65, 2),
				MintOut:         sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(400000)),
			},
			// interestOfPeriod = 6_000000 * 4 * (1-0) / (10*60*24*365) = 5
			expPass: true,
			expRes: &types.EstimateMintByCollateralInResponse{
				CollateralIn: sdk.NewCoin(suite.bcDenom, sdk.ZeroInt()),            // 1e7 > (6e6 + 5 + 4e5 / (1-0.01)) / 0.65 / 0.99 = 9951897
				LionIn:       sdk.NewCoin(merlion.AttoLionDenom, sdk.ZeroInt()),    // 0.5 + (3e15*1e-10) / (1e7*0.99) / 0.05 * (0.8-0.5) = 0.682 > 0.65
				MintFee:      sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(4040)), // 4e5 / (1-0.01) * 0.01
			},
		},
		{
			name: "correct",
			req: &types.EstimateMintByCollateralInRequest{
				CollateralDenom: suite.bcDenom,
				Ltv:             sdk.NewDecWithPrec(7, 1),
				MintOut:         sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(1_000000)),
			},
			expPass: true,
			expRes: &types.EstimateMintByCollateralInResponse{
				CollateralIn: sdk.NewCoin(suite.bcDenom, sdk.NewInt(115593)),                    // (6e6 + 5 + 1e6/(1-0.01)) / 0.7 / 0.99 - 1e7
				LionIn:       sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(338_145690_000000)), // (0.7-0.5) / (0.8-0.5) * 0.05 * (1e7+115593)*0.99 / 1e-10 - 3e15
				MintFee:      sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(10101)),             // 1e6 / 0.99 * 0.01
			},
		},
		{
			name: "collateral over ceiling",
			req: &types.EstimateMintByCollateralInRequest{
				CollateralDenom: suite.bcDenom,
				Ltv:             sdk.NewDecWithPrec(5, 1),
				MintOut:         sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(1_000000)),
			},
			// CollateralIn = (6e6 + 5 + 1e6/(1-0.01)) / 0.5 / 0.99 - 1e7 = 4_161830
			// pool.Collateral max adding:  1_000000
			expPass: false,
			expErr:  types.ErrCollateralCeiling,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			suite.setupEstimationTest()
			if tc.malleate != nil {
				tc.malleate()
			}

			ctx := sdk.WrapSDKContext(suite.ctx)
			tc.req.Account = suite.accAddress.String()
			res, err := suite.queryClient.EstimateMintByCollateralIn(ctx, tc.req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expRes, res)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.expErr)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestEstimateMintByCollateralOut() {
	testCases := []struct {
		name     string
		malleate func()
		req      *types.EstimateMintByCollateralOutRequest
		expPass  bool
		expErr   error
		expRes   *types.EstimateMintByCollateralOutResponse
	}{
		// {
		//	name: "mer price too low",
		//	malleate: func() {
		//		suite.app.OracleKeeper.SetExchangeRate(suite.ctx, merlion.MicroUSMDenom, sdk.NewDecWithPrec(989, 3))
		//	},
		//	req: &types.EstimateMintByCollateralOutRequest{
		//		CollateralIn: sdk.NewCoin(suite.bcDenom, sdk.ZeroInt()),
		//	},
		//	expPass: false,
		//	expErr:  types.ErrMerPriceTooLow,
		// },
		{
			name: "collateral denom not found",
			req: &types.EstimateMintByCollateralOutRequest{
				CollateralIn: sdk.NewCoin("fil", sdk.ZeroInt()),
			},
			expPass: false,
			expErr:  types.ErrCollateralCoinNotFound,
		},
		{
			name: "collateral denom disabled",
			req: &types.EstimateMintByCollateralOutRequest{
				CollateralIn: sdk.NewCoin("eth", sdk.ZeroInt()),
			},
			expPass: false,
			expErr:  types.ErrCollateralCoinDisabled,
		},
		{
			name: "ltv too low",
			req: &types.EstimateMintByCollateralOutRequest{
				CollateralIn: sdk.NewCoin(suite.bcDenom, sdk.ZeroInt()),
				Ltv:          sdk.NewDecWithPrec(4, 1),
			},
			expPass: false,
			expErr:  types.ErrLTVOutOfRange,
		},
		{
			name: "ltv too high",
			req: &types.EstimateMintByCollateralOutRequest{
				CollateralIn: sdk.NewCoin(suite.bcDenom, sdk.ZeroInt()),
				Ltv:          sdk.NewDecWithPrec(9, 1),
			},
			expPass: false,
			expErr:  types.ErrLTVOutOfRange,
		},
		{
			name: "collateral over ceiling",
			req: &types.EstimateMintByCollateralOutRequest{
				CollateralIn: sdk.NewCoin(suite.bcDenom, sdk.NewInt(1_000001)),
				Ltv:          sdk.NewDecWithPrec(5, 1),
			},
			// pool.Collateral max adding:  1_000000
			expPass: false,
			expErr:  types.ErrCollateralCeiling,
		},
		{
			name: "mer over ceiling",
			req: &types.EstimateMintByCollateralOutRequest{
				CollateralIn: sdk.NewCoin(suite.bcDenom, sdk.NewInt(1e6)),
				Ltv:          sdk.NewDecWithPrec(8, 1),
			},
			// interestOfPeriod = 6_000000 * 4 * (1-0) / (10*60*24*365) = 5
			// (1e7+1e6)*0.99*0.8-(6e6+5) + 8e6 = 10_711995 > 1e7
			expPass: false,
			expErr:  types.ErrMerCeiling,
		},
		{
			name: "correct",
			malleate: func() {
				poolColl, found := suite.app.MakerKeeper.GetPoolCollateral(suite.ctx, suite.bcDenom)
				suite.Require().True(found)
				poolColl.Collateral.Amount = sdk.NewInt(1_000000)
				suite.app.MakerKeeper.SetPoolCollateral(suite.ctx, poolColl)
			},
			req: &types.EstimateMintByCollateralOutRequest{
				CollateralIn: sdk.NewCoin(suite.bcDenom, sdk.NewInt(1e6)),
				Ltv:          sdk.NewDecWithPrec(7, 1),
			},
			expPass: true,
			expRes: &types.EstimateMintByCollateralOutResponse{
				LionIn:  sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(63e13)),    // (0.7-0.5) / (0.8-0.5) * 0.05 * (1e7+1e6)*0.99 / 1e-10 - 3e15
				MintOut: sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(1_606765)), // ((1e7+1e6)*0.99*0.7 - (6e6+5)) * (1-0.01)
				MintFee: sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(16230)),    // ((1e7+1e6)*0.99*0.7 - (6e6+5)) * 0.01
			},
		},
		{
			name: "correct all zero",
			req: &types.EstimateMintByCollateralOutRequest{
				CollateralIn: sdk.NewCoin(suite.bcDenom, sdk.ZeroInt()),
				Ltv:          sdk.NewDecWithPrec(5, 1),
			},
			expPass: true,
			expRes: &types.EstimateMintByCollateralOutResponse{
				LionIn:  sdk.NewCoin(merlion.AttoLionDenom, sdk.ZeroInt()),
				MintOut: sdk.NewCoin(merlion.MicroUSMDenom, sdk.ZeroInt()),
				MintFee: sdk.NewCoin(merlion.MicroUSMDenom, sdk.ZeroInt()),
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			suite.setupEstimationTest()
			if tc.malleate != nil {
				tc.malleate()
			}

			ctx := sdk.WrapSDKContext(suite.ctx)
			tc.req.Account = suite.accAddress.String()
			res, err := suite.queryClient.EstimateMintByCollateralOut(ctx, tc.req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expRes, res)
			} else {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.expErr)
			}
		})
	}

}

func (suite *KeeperTestSuite) setupEstimationTest() {
	// set prices
	suite.app.OracleKeeper.SetExchangeRate(suite.ctx, suite.bcDenom, sdk.NewDecWithPrec(99, 2))
	suite.app.OracleKeeper.SetExchangeRate(suite.ctx, "eth", sdk.NewDec(1000_000000))
	suite.app.OracleKeeper.SetExchangeRate(suite.ctx, "fil", sdk.NewDec(5_000000))
	suite.app.OracleKeeper.SetExchangeRate(suite.ctx, merlion.AttoLionDenom, sdk.NewDecWithPrec(100, 12))
	suite.app.OracleKeeper.SetExchangeRate(suite.ctx, merlion.MicroUSMDenom, sdk.NewDecWithPrec(101, 2))

	// set risk params
	brp, brp2 := suite.dummyBackingRiskParams()
	suite.app.MakerKeeper.SetBackingRiskParams(suite.ctx, brp)
	suite.app.MakerKeeper.SetBackingRiskParams(suite.ctx, brp2)

	crp, crp2 := suite.dummyCollateralRiskParams()
	suite.app.MakerKeeper.SetCollateralRiskParams(suite.ctx, crp)
	suite.app.MakerKeeper.SetCollateralRiskParams(suite.ctx, crp2)

	// set pool and total backing
	suite.app.MakerKeeper.SetPoolBacking(suite.ctx, types.PoolBacking{
		MerMinted:  sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(8_000000)),
		Backing:    sdk.NewCoin(suite.bcDenom, sdk.NewInt(9_000000)),
		LionBurned: sdk.NewCoin(merlion.AttoLionDenom, sdk.ZeroInt()),
	})
	suite.app.MakerKeeper.SetTotalBacking(suite.ctx, types.TotalBacking{
		MerMinted:  sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(8_500000)),
		LionBurned: sdk.NewCoin(merlion.AttoLionDenom, sdk.ZeroInt()),
	})

	// set account, pool and total collateral
	suite.app.MakerKeeper.SetAccountCollateral(suite.ctx, suite.accAddress, types.AccountCollateral{
		Account:             suite.accAddress.String(),
		Collateral:          sdk.NewCoin(suite.bcDenom, sdk.NewInt(10_000000)),
		MerDebt:             sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(6_000000)),
		LionCollateralized:  sdk.NewCoin(merlion.AttoLionDenom, sdk.NewInt(3e15)),
		LastInterest:        sdk.NewCoin(merlion.MicroUSMDenom, sdk.ZeroInt()),
		LastSettlementBlock: 0,
	})
	suite.app.MakerKeeper.SetPoolCollateral(suite.ctx, types.PoolCollateral{
		Collateral:         sdk.NewCoin(suite.bcDenom, sdk.NewInt(15_000000)),
		MerDebt:            sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(8_000000)),
		LionCollateralized: sdk.NewCoin(merlion.AttoLionDenom, sdk.ZeroInt()),
	})
	suite.app.MakerKeeper.SetTotalCollateral(suite.ctx, types.TotalCollateral{
		MerDebt:            sdk.NewCoin(merlion.MicroUSMDenom, sdk.NewInt(10_000000)),
		LionCollateralized: sdk.NewCoin(merlion.AttoLionDenom, sdk.ZeroInt()),
	})
}
