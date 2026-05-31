package llm

import "testing"

func TestChatCompletionsURL(t *testing.T) {
	if u := chatCompletionsURL(ArkInferenceBase); u != ArkInferenceBase+"/chat/completions" {
		t.Fatalf("got %s", u)
	}
	if u := chatCompletionsURL(ArkCodingBase + "/chat/completions"); u != ArkCodingBase+"/chat/completions" {
		t.Fatalf("got %s", u)
	}
}

func TestValidateVolcengineModelAuto(t *testing.T) {
	if got := validateVolcengineModel(ArkInferenceBase, "auto"); got == "" {
		t.Fatal("expected hint for auto")
	}
}
