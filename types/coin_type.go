package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	//1 iris = 10^3 iris-milli
	Milli      = "milli"
	MilliScale = 3

	//1 iris = 10^6 iris-micro
	Micro      = "micro"
	MicroScale = 6

	//1 iris = 10^9 iris-nano
	Nano      = "nano"
	NanoScale = 9

	//1 iris = 10^12 iris-pico
	Pico      = "pico"
	PicoScale = 12

	//1 iris = 10^15 iris-femto
	Femto      = "femto"
	FemtoScale = 15

	//1 iris = 10^18 iris-atto
	Atto      = "atto"
	AttoScale = 18

	MinDenomSuffix = "-min"
)

var (
	IrisCoinType    = NewIrisCoinType()
	AttoScaleFactor = IrisCoinType.MinUnit.GetScaleFactor()
)

type Unit struct {
	Denom   string `json:"denom"`
	Decimal uint8  `json:"decimal"`
}

func (u Unit) String() string {
	return fmt.Sprintf("%s: %d",
		u.Denom, u.Decimal,
	)
}

func NewUnit(denom string, decimal uint8) Unit {
	return Unit{
		Denom:   denom,
		Decimal: decimal,
	}
}

func (u Unit) GetScaleFactor() sdk.Int {
	return sdk.NewIntWithDecimal(1, int(u.Decimal))
}

type Units []Unit

func (u Units) String() (out string) {
	for _, val := range u {
		out += val.String() + ",  "
	}
	if len(out) > 3 {
		out = out[:len(out)-3]
	}
	return
}

type CoinType struct {
	Name    string `json:"name"`
	MinUnit Unit   `json:"min_unit"`
	Units   Units  `json:"units"`
	Desc    string `json:"desc"`
}

func (ct CoinType) Convert(srcCoinStr string, destDenom string) (destCoinStr string, err error) {
	coin, err := sdk.ParseCoin(srcCoinStr)
	if err != nil {
		return destCoinStr, err
	}

	destUnit, err := ct.GetUnit(destDenom)
	if err != nil {
		return destCoinStr, errors.New("destination unit (%s) not defined" + destDenom)
	}

	srcUnit, err := ct.GetUnit(coin.Denom)
	if err != nil {
		return destCoinStr, errors.New("source unit (%s) not defined" + coin.Denom)
	}
	if srcUnit.Denom == destDenom {
		return srcCoinStr, nil
	}
	// dest amount = src amount * (10^(dest scale) / 10^(src scale))
	ratScale := sdk.NewDecFromInt(destUnit.GetScaleFactor())
	srcScale := sdk.NewDecFromInt(srcUnit.GetScaleFactor())
	amount := sdk.NewDecFromInt(coin.Amount) // convert src amount to dest unit

	amt := amount.Mul(ratScale).Quo(srcScale).String()
	destCoinStr = fmt.Sprintf("%s%s", amt, destUnit.Denom)
	return destCoinStr, nil
}

func (ct CoinType) ConvertToMinDenomCoin(srcCoinStr string) (coin sdk.Coin, err error) {
	destCoinStr, err := ct.Convert(srcCoinStr, ct.MinUnit.Denom)
	if err == nil {
		return coin, errors.New("convert error")
	}
	return sdk.ParseCoin(destCoinStr)
}

func (ct CoinType) GetUnit(denom string) (u Unit, err error) {
	for _, unit := range ct.Units {
		if strings.ToLower(denom) == strings.ToLower(unit.Denom) {
			return unit, nil
		}
	}
	return u, errors.New("unit (%s) not found" + denom)
}

func (ct CoinType) GetMainUnit() (unit Unit) {
	unit, _ = ct.GetUnit(ct.Name)
	return unit
}

func (ct CoinType) String() string {
	return fmt.Sprintf(`CoinType:
 Name:     %s
 MinUnit:  %s
 Units:    %s
 Desc:     %s`,
		ct.Name, ct.MinUnit, ct.Units, ct.Desc,
	)
}

func NewIrisCoinType() CoinType {
	units := make(Units, 7)

	units[0] = NewUnit(Iris, 0)
	units[1] = NewUnit(fmt.Sprintf("%s-%s", Iris, Milli), MilliScale)
	units[2] = NewUnit(fmt.Sprintf("%s-%s", Iris, Micro), MicroScale)
	units[3] = NewUnit(fmt.Sprintf("%s-%s", Iris, Nano), NanoScale)
	units[4] = NewUnit(fmt.Sprintf("%s-%s", Iris, Pico), PicoScale)
	units[5] = NewUnit(fmt.Sprintf("%s-%s", Iris, Femto), FemtoScale)
	units[6] = NewUnit(fmt.Sprintf("%s-%s", Iris, Atto), AttoScale)

	return CoinType{
		Name:    Iris,
		Units:   units,
		MinUnit: units[6],
		Desc:    "IRIS Network",
	}
}

func GetCoinNameByDenom(denom string) (coinName string, err error) {
	denom = strings.ToLower(denom)
	if strings.HasPrefix(denom, Iris+"-") {
		if _, err := IrisCoinType.GetUnit(denom); err != nil {
			return "", fmt.Errorf("invalid denom for getting coin name: %s", denom)
		}
		return Iris, nil
	}
	if !IsCoinMinDenomValid(denom) {
		return "", fmt.Errorf("invalid denom for getting coin name: %s", denom)
	}
	coinName = strings.TrimSuffix(denom, MinDenomSuffix)
	if coinName == "" {
		return coinName, fmt.Errorf("coin name is empty")
	}
	return coinName, nil
}

func GetCoinMinDenom(coinName string) (denom string, err error) {
	coinName = strings.ToLower(strings.TrimSpace(coinName))

	if coinName == Iris {
		return IrisAtto, nil
	}

	return fmt.Sprintf("%s%s", coinName, MinDenomSuffix), nil
}
