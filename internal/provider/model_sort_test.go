package provider

import (
	"slices"
	"testing"
)

// A user with one model on several providers (a relay, a VPS's magpie)
// asked for the agent's lists to put the models whose names agree together
// instead of each provider's models together (ModelSort, SortName). The
// routing groups keep magpie's own place ahead of them, and the models the
// user dragged keep the place they put them in.
func TestModelSortByNamePutsOneModelsProvidersTogether(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	for _, p := range []Provider{
		{ID: "relay", Name: "Relay", Key: "k", Chat: "http://127.0.0.1:1/v1", Models: []string{"glm-5.3", "kimi-k3"}},
		{ID: "vps", Name: "VPS", Key: "k", Chat: "http://127.0.0.1:1/v1", Models: []string{"z-ai/GLM-5.3", "moonshotai/Kimi-K3"}},
	} {
		if err := Save(p); err != nil {
			t.Fatal(err)
		}
	}
	// the names the two vendors' lists give them: one model under two
	// providers, each spelling it its own way
	for id, name := range map[string]string{
		"relay/glm-5.3":          "GLM-5.3",
		"relay/kimi-k3":          "Kimi K3",
		"vps/z-ai/GLM-5.3":       "GLM-5.3",
		"vps/moonshotai/Kimi-K3": "Kimi K3",
	} {
		if err := SetModelName(id, name); err != nil {
			t.Fatal(err)
		}
	}
	ids := func(agent string) []string {
		shown, _ := CatalogFor(agent)
		var out []string
		for _, e := range shown {
			out = append(out, e.ID)
		}
		return out
	}
	// magpie's own order: the groups magpie found lead (one model on two
	// providers is a group of them), then each provider's models together
	byProvider := []string{"group/auto-glm-5-3", "group/auto-kimi-k3", "relay/glm-5.3", "relay/kimi-k3", "vps/z-ai/GLM-5.3", "vps/moonshotai/Kimi-K3"}
	if got, want := ids("opencode"), byProvider; !slices.Equal(got, want) {
		t.Fatalf("by provider %v, want %v", got, want)
	}
	if err := SetModelSort(SortName); err != nil {
		t.Fatal(err)
	}
	if ModelSort() != SortName {
		t.Fatal("SortName not kept by the save")
	}
	// by name: the two spellings of one model sit together, names in
	// order; the groups keep their place ahead of the models
	if got, want := ids("opencode"), []string{"group/auto-glm-5-3", "group/auto-kimi-k3", "relay/glm-5.3", "vps/z-ai/GLM-5.3", "relay/kimi-k3", "vps/moonshotai/Kimi-K3"}; !slices.Equal(got, want) {
		t.Fatalf("by name %v, want %v", got, want)
	}
	// every agent's list is put by name: the setting is one for them all
	if got, want := ids("codex"), []string{"group/auto-glm-5-3", "group/auto-kimi-k3", "relay/glm-5.3", "vps/z-ai/GLM-5.3", "relay/kimi-k3", "vps/moonshotai/Kimi-K3"}; !slices.Equal(got, want) {
		t.Fatalf("codex %v, want %v", got, want)
	}
	// the user's own order still comes first
	if err := SetModelOrder("opencode", []string{"vps/moonshotai/Kimi-K3"}); err != nil {
		t.Fatal(err)
	}
	if got, want := ids("opencode"), []string{"vps/moonshotai/Kimi-K3", "group/auto-glm-5-3", "group/auto-kimi-k3", "relay/glm-5.3", "vps/z-ai/GLM-5.3", "relay/kimi-k3"}; !slices.Equal(got, want) {
		t.Fatalf("with a dragged model %v, want %v", got, want)
	}
	// back to magpie's own order
	if err := SetModelSort(SortProvider); err != nil {
		t.Fatal(err)
	}
	if ModelSort() != SortProvider {
		t.Fatal("SortProvider not kept by the save")
	}
	if got, want := ids("opencode"), []string{"vps/moonshotai/Kimi-K3", "group/auto-glm-5-3", "group/auto-kimi-k3", "relay/glm-5.3", "relay/kimi-k3", "vps/z-ai/GLM-5.3"}; !slices.Equal(got, want) {
		t.Fatalf("back to by provider %v, want %v", got, want)
	}
}

// A routing group is not a model of a name: by name, the groups keep
// magpie's own place ahead of the models rather than sitting among them.
func TestModelSortByNameKeepsGroupsFirst(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := Save(Provider{ID: "relay", Name: "Relay", Key: "k", Chat: "http://127.0.0.1:1/v1", Models: []string{"glm-5.3", "zzz-model"}}); err != nil {
		t.Fatal(err)
	}
	if err := SaveGroup(Group{Name: "Mix", Members: []string{"relay/zzz-model"}}); err != nil {
		t.Fatal(err)
	}
	if err := SetModelSort(SortName); err != nil {
		t.Fatal(err)
	}
	shown, _ := CatalogFor("opencode")
	var got []string
	for _, e := range shown {
		got = append(got, e.ID)
	}
	// the group leads, then the models by name
	if want := []string{"group/mix", "relay/glm-5.3", "relay/zzz-model"}; !slices.Equal(got, want) {
		t.Fatalf("%v, want %v", got, want)
	}
}
