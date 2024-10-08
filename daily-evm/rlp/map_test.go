package rlp

import (
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"testing"

	"github.com/dailycrypto-me/daily-evm/daily/util/asserts"
	"github.com/dailycrypto-me/daily-evm/daily/util/bin"
)

func TestMap(t *testing.T) {
	fmt.Println(t.Name())
	h := fnv.New64a()
	h.Write(bin.BytesView(t.Name()))
	rnd := rand.New(rand.NewSource(int64(h.Sum64())))
	type TestStruct0 struct {
		F0 uint32
		F1 [23]byte
	}
	type TestMap0 = map[string][]TestStruct0
	type TestStruct1 struct {
		F0 uint64
		F1 string
		F2 TestMap0
	}
	type TestMap1 = map[string]TestStruct1
	expected := make(TestMap1)
	for i, n := 0, rnd.Intn(15); i < n; i++ {
		var el1 TestStruct1
		el1.F0 = rnd.Uint64()
		el1.F1 = string(bin.RandomBytes(rnd.Intn(15), rnd))
		el1.F2 = make(TestMap0)
		for i, n := 0, rnd.Intn(15); i < n; i++ {
			var el0_list []TestStruct0
			for i, n := 0, rnd.Intn(15); i < n; i++ {
				var el0 TestStruct0
				el0.F0 = rnd.Uint32()
				for i := 0; i < len(el0.F1); i++ {
					el0.F1[i] = byte(rnd.Int() % math.MaxUint8)
				}
				el0_list = append(el0_list, el0)
			}
			el1.F2[string(bin.RandomBytes(10, rnd))] = el0_list
		}
		expected[string(bin.RandomBytes(rnd.Intn(15), rnd))] = el1
	}
	actual := make(TestMap1)
	MustDecodeBytes(MustEncodeToBytes(expected), &actual)
	asserts.Holds(len(actual) == len(expected))
	for k, expected := range expected {
		actual := actual[k]
		asserts.Holds(expected.F0 == actual.F0)
		asserts.Holds(expected.F1 == actual.F1)
		asserts.Holds(len(expected.F2) == len(actual.F2))
		for k, expected := range expected.F2 {
			actual := actual.F2[k]
			asserts.Holds(len(expected) == len(actual))
			for i, expected := range expected {
				actual := actual[i]
				asserts.Holds(expected.F0 == actual.F0)
				asserts.Holds(expected.F1 == actual.F1)
			}
		}
	}
}
