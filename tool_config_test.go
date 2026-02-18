package main

import "testing"

func TestOpenAIBaseURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty uses canonical default", input: "", want: "https://api.a21e.com/v1"},
		{name: "appends version path", input: "https://api.a21e.com", want: "https://api.a21e.com/v1"},
		{name: "keeps explicit version path", input: "https://api.a21e.com/v1", want: "https://api.a21e.com/v1"},
		{name: "trims slash and keeps version", input: "https://api.a21e.com/v1/", want: "https://api.a21e.com/v1"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := openAIBaseURL(tc.input)
			if got != tc.want {
				t.Fatalf("openAIBaseURL(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestUpsertManagedBlock(t *testing.T) {
	t.Parallel()

	const start = "# >>> a21e openai_cli_custom >>>"
	const end = "# <<< a21e openai_cli_custom <<<"
	const block1 = "# >>> a21e openai_cli_custom >>>\nline-1\n# <<< a21e openai_cli_custom <<<"
	const block2 = "# >>> a21e openai_cli_custom >>>\nline-2\n# <<< a21e openai_cli_custom <<<"

	inserted, changed, err := upsertManagedBlock("export PATH=\"$HOME/.local/bin:$PATH\"\n", start, end, block1)
	if err != nil {
		t.Fatalf("insert returned unexpected error: %v", err)
	}
	if !changed {
		t.Fatalf("expected insert to report changed=true")
	}
	if inserted == "" || inserted == "export PATH=\"$HOME/.local/bin:$PATH\"\n" {
		t.Fatalf("expected managed block to be inserted")
	}

	replaced, changed, err := upsertManagedBlock(inserted, start, end, block2)
	if err != nil {
		t.Fatalf("replace returned unexpected error: %v", err)
	}
	if !changed {
		t.Fatalf("expected replace to report changed=true")
	}
	if replaced == inserted {
		t.Fatalf("expected block replacement to change content")
	}

	stable, changed, err := upsertManagedBlock(replaced, start, end, block2)
	if err != nil {
		t.Fatalf("stable replace returned unexpected error: %v", err)
	}
	if changed {
		t.Fatalf("expected replacing with identical block to report changed=false")
	}
	if stable != replaced {
		t.Fatalf("expected stable output when block is unchanged")
	}
}

func TestMergeA21ESettings(t *testing.T) {
	t.Parallel()

	base := []byte("{\n  \"editor.fontSize\": 14\n}\n")
	updated, changed, err := mergeA21ESettings(base, "a21e_test_key", "https://api.a21e.com")
	if err != nil {
		t.Fatalf("merge returned unexpected error: %v", err)
	}
	if !changed {
		t.Fatalf("expected merge to report changed=true for first update")
	}

	updatedAgain, changedAgain, err := mergeA21ESettings(updated, "a21e_test_key", "https://api.a21e.com")
	if err != nil {
		t.Fatalf("second merge returned unexpected error: %v", err)
	}
	if changedAgain {
		t.Fatalf("expected second merge to report changed=false")
	}
	if string(updatedAgain) != string(updated) {
		t.Fatalf("expected second merge output to remain unchanged")
	}
}
