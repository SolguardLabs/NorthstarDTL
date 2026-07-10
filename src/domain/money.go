package domain

import "fmt"

const (
	RateScale = int64(1_000_000)
	BpsScale  = int64(10_000)
)

type Money int64

func (m Money) Int64() int64 { return int64(m) }

func (m Money) String() string {
	return fmt.Sprintf("%d", m)
}

func (m Money) Positive() bool {
	return m > 0
}

func (m Money) NonNegative() bool {
	return m >= 0
}

func (m Money) Add(other Money) Money {
	return m + other
}

func (m Money) Sub(other Money) Money {
	return m - other
}

func (m Money) MulRate(ratePpm int64) Money {
	if ratePpm < 0 {
		return 0
	}
	return Money((int64(m) * ratePpm) / RateScale)
}

func (m Money) MulBps(bps int64) Money {
	if bps <= 0 {
		return 0
	}
	return Money((int64(m) * bps) / BpsScale)
}

func (m Money) ClampFloor(floor Money) Money {
	if m < floor {
		return floor
	}
	return m
}

func (m Money) ClampCeiling(ceiling Money) Money {
	if m > ceiling {
		return ceiling
	}
	return m
}

func MinMoney(a, b Money) Money {
	if a < b {
		return a
	}
	return b
}

func MaxMoney(a, b Money) Money {
	if a > b {
		return a
	}
	return b
}
