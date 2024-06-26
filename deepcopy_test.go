package deepcopy_test

import (
	"fmt"
	. "reflect"
	"testing"
	"time"

	"github.com/cookieai-jar/go-deepcopy"
)

func ExampleAnything() {
	tests := []interface{}{
		`"Now cut that out!"`,
		39,
		true,
		false,
		2.14,
		[]string{
			"Phil Harris",
			"Rochester van Jones",
			"Mary Livingstone",
			"Dennis Day",
		},
		[2]string{
			"Jell-O",
			"Grape-Nuts",
		},
	}

	for _, expected := range tests {
		actual := deepcopy.MustAnything(expected)
		fmt.Println(actual)
	}
	// Output:
	// "Now cut that out!"
	// 39
	// true
	// false
	// 2.14
	// [Phil Harris Rochester van Jones Mary Livingstone Dennis Day]
	// [Jell-O Grape-Nuts]
}

type Foo struct {
	Foo *Foo
	Bar int
}

func ExampleMap() {
	x := map[string]*Foo{
		"foo": &Foo{Bar: 1},
		"bar": &Foo{Bar: 2},
	}
	y := deepcopy.MustAnything(x).(map[string]*Foo)
	for _, k := range []string{"foo", "bar"} { // to ensure consistent order
		fmt.Printf("x[\"%v\"] = y[\"%v\"]: %v\n", k, k, x[k] == y[k])
		fmt.Printf("x[\"%v\"].Foo = y[\"%v\"].Foo: %v\n", k, k, x[k].Foo == y[k].Foo)
		fmt.Printf("x[\"%v\"].Bar = y[\"%v\"].Bar: %v\n", k, k, x[k].Bar == y[k].Bar)
	}
	// Output:
	// x["foo"] = y["foo"]: false
	// x["foo"].Foo = y["foo"].Foo: true
	// x["foo"].Bar = y["foo"].Bar: true
	// x["bar"] = y["bar"]: false
	// x["bar"].Foo = y["bar"].Foo: true
	// x["bar"].Bar = y["bar"].Bar: true
}

func TestInterface(t *testing.T) {
	x := []interface{}{nil}
	y := deepcopy.MustAnything(x).([]interface{})
	if !DeepEqual(x, y) || len(y) != 1 {
		t.Errorf("expect %v == %v; y had length %v (expected 1)", x, y, len(y))
	}
	var a interface{}
	b := deepcopy.MustAnything(a)
	if a != b {
		t.Errorf("expected %v == %v", a, b)
	}
}

func ExampleAvoidInfiniteLoops() {
	x := &Foo{
		Bar: 4,
	}
	x.Foo = x
	y := deepcopy.MustAnything(x).(*Foo)
	fmt.Printf("x == y: %v\n", x == y)
	fmt.Printf("x == x.Foo: %v\n", x == x.Foo)
	fmt.Printf("y == y.Foo: %v\n", y == y.Foo)
	// Output:
	// x == y: false
	// x == x.Foo: true
	// y == y.Foo: true
}

func TestUnsupportedKind(t *testing.T) {
	x := func() {}

	tests := []interface{}{
		x,
		map[bool]interface{}{true: x},
		[]interface{}{x},
	}

	for _, test := range tests {
		y, err := deepcopy.Anything(test)
		if y != nil {
			t.Errorf("expected %v to be nil", y)
		}
		if err == nil {
			t.Errorf("expected err to not be nil")
		}
	}
}

func TestUnsupportedKindPanicsOnMust(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected a panic; didn't get one")
		}
	}()
	x := func() {}
	deepcopy.MustAnything(x)
}

func TestMismatchedTypesFail(t *testing.T) {
	tests := []struct {
		input interface{}
		kind  Kind
	}{
		{
			map[int]int{1: 2, 2: 4, 3: 8},
			Map,
		},
		{
			[]int{2, 8},
			Slice,
		},
	}
	for _, test := range tests {
		for kind, copier := range deepcopy.Copiers {
			if kind == test.kind {
				continue
			}
			actual, err := copier(test.input, nil)
			if actual != nil {

				t.Errorf("%v attempted value %v as %v; should be nil value, got %v", test.kind, test.input, kind, actual)
			}
			if err == nil {
				t.Errorf("%v attempted value %v as %v; should have gotten an error", test.kind, test.input, kind)
			}
		}
	}
}

func TestCopyNilValue(t *testing.T) {
	type Bar struct {
		Baz string
	}
	type Foo struct {
		Bar *Bar
	}
	s := &Foo{
		Bar: nil,
	}
	c, err := deepcopy.Anything(s)
	if err != nil {
		t.Fatalf("deepcopy of struct %#v failed", s)
	}

	if !DeepEqual(s, c) {
		t.Fatalf("original and copied struct are not equal: %+v != %+v", s, c)
	}
}

func TestTimeType(t *testing.T) {
	src := time.Date(2016, 1, 1, 1, 0, 0, 0, time.UTC)
	dst, err := deepcopy.Anything(src)
	if err != nil {
		t.Errorf("expected no error; got %v", err)
	}
	resultTime, ok := dst.(time.Time)
	if !ok {
		t.Errorf("expected a time.Time; got %v", resultTime)
	}
	if !DeepEqual(src, dst) {
		t.Errorf("expect %v == %v; ", src, dst)
	}

}

func TestTimePtrType(t *testing.T) {
	type Foo struct {
		T    time.Time
		TPtr *time.Time
	}

	aTime := time.Date(2016, 1, 1, 1, 0, 0, 0, time.UTC)
	anotherTime := aTime.Add(24 * time.Hour)
	src := Foo{
		T:    aTime,
		TPtr: &anotherTime,
	}
	dst, err := deepcopy.Anything(src)
	if err != nil {
		t.Errorf("expected no error; got %v", err)
	}
	res, ok := dst.(Foo)
	if !ok {
		t.Errorf("expected a time.Time; got %v", res)
	}
	if !DeepEqual(src, dst) {
		t.Errorf("expect %v == %v; ", src, dst)
	}
}
