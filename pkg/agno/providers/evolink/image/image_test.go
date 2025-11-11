package image

import "testing"

func TestOptionsValidate(t *testing.T) {
	// valid
	o := Options{Model: ModelGPT4O, Prompt: "p", Size: "1:1", N: 2}
	if err := o.validate(); err != nil {
		t.Fatalf("valid options: %v", err)
	}

	// invalid size
	o2 := Options{Model: ModelGPT4O, Prompt: "p", Size: "4:3", N: 1}
	if err := o2.validate(); err == nil {
		t.Fatal("expected invalid size error")
	}

	// mask requires exactly one reference and .png
	o3 := Options{Model: ModelGPT4O, Prompt: "p", Size: "1:1", N: 1, MaskURL: "https://x/mask.jpg", References: []string{"a"}}
	if err := o3.validate(); err == nil {
		t.Fatal("expected mask png error")
	}

	o4 := Options{Model: ModelGPT4O, Prompt: "p", Size: "1:1", N: 1, MaskURL: "https://x/mask.png"}
	if err := o4.validate(); err == nil {
		t.Fatal("expected exactly one reference error")
	}

	// invalid model
	o5 := Options{Model: Model("unknown"), Prompt: "p", Size: "1:1", N: 1}
	if err := o5.validate(); err == nil {
		t.Fatal("expected unsupported model error")
	}
}
