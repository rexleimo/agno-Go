package video

import (
	"reflect"
	"testing"
)

func TestOptionsValidateSoraDefaults(t *testing.T) {
	opts := Options{Model: ModelSora2, Prompt: "p"}
	if err := opts.validate(); err != nil {
		t.Fatalf("validate sora: %v", err)
	}
	if opts.AspectRatio != "16:9" {
		t.Fatalf("expected aspect_ratio default 16:9, got %s", opts.AspectRatio)
	}
	if opts.DurationSeconds != 10 {
		t.Fatalf("expected duration 10, got %d", opts.DurationSeconds)
	}
	if opts.RemoveWatermark == nil || !*opts.RemoveWatermark {
		t.Fatalf("expected remove_watermark default true")
	}
	p := opts.buildPayload()
	if _, ok := p["remove_watermark"]; !ok {
		t.Fatalf("payload missing remove_watermark")
	}
}

func TestOptionsValidateWanImageToVideoRequiresImage(t *testing.T) {
	opts := Options{Model: ModelWan25ImageToVideo, Prompt: "p"}
	if err := opts.validate(); err == nil {
		t.Fatal("expected error when image_urls missing")
	}
	opts.ImageURLs = []string{"https://example.com/img.png"}
	opts.DurationSeconds = 5
	if err := opts.validate(); err != nil {
		t.Fatalf("validate wan image: %v", err)
	}
	p := opts.buildPayload()
	urls, ok := p["image_urls"].([]string)
	if !ok || len(urls) != 1 {
		t.Fatalf("payload must include image_urls: %#v", p)
	}
}

func TestOptionsValidateSeedanceDefaults(t *testing.T) {
	opts := Options{Model: ModelSeedance10ProFast, Prompt: "p", ImageURLs: []string{"https://ex.com/seed.png"}}
	opts.DurationSeconds = 6
	if err := opts.validate(); err != nil {
		t.Fatalf("validate seedance: %v", err)
	}
	if opts.AspectRatio != "adaptive" {
		t.Fatalf("expected adaptive aspect ratio, got %s", opts.AspectRatio)
	}
	p := opts.buildPayload()
	if !reflect.DeepEqual(p["image_urls"], opts.ImageURLs) {
		t.Fatalf("payload image_urls mismatch: %#v", p["image_urls"])
	}
}
