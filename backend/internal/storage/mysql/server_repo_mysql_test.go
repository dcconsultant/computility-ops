package mysql

import "testing"

func TestServerMetadataModelCodes_PrefersServerBeforeLegacyTypo(t *testing.T) {
	got := serverMetadataModelCodes()
	if len(got) != 2 {
		t.Fatalf("len(serverMetadataModelCodes())=%d, want 2", len(got))
	}
	if got[0] != "server" || got[1] != "sever" {
		t.Fatalf("serverMetadataModelCodes()=%v, want [server sever]", got)
	}
}
