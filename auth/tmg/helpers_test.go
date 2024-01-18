package tmg

import (
	"context"
	"testing"
)

func TestHelpersEdgeCases(t *testing.T) {

	t.Run("GetAuth/EmptySiteURL", func(t *testing.T) {
		cnfg := &AuthCnfg{SiteURL: ""}
		if _, _, err := GetAuth(context.Background(), cnfg); err == nil {
			t.Error("empty SiteURL should not go")
		}
	})

}
