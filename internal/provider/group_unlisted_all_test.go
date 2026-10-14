package provider

import (
	"slices"
	"testing"
)

// 蓝猫 on Discord: a group's Model lists pick ("The group alone") was one
// group at a time. GroupsUnlisted does it for every group at once: the
// models each group names, or its patterns match, leave the lists agents
// and /v1/models show, so each group is offered alone. Each group's own
// pick is kept, not written over, so switching the whole-list setting off
// leaves the groups as the user had them.
func TestEveryGroupUnlistedAtOnce(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	for _, p := range []Provider{
		{ID: "a", Name: "a", Key: "k", Chat: "http://127.0.0.1:1/v1", Models: []string{"m1", "m2"}},
		{ID: "b", Name: "b", Key: "k", Chat: "http://127.0.0.1:1/v1", Models: []string{"n1"}},
	} {
		if err := Save(p); err != nil {
			t.Fatal(err)
		}
	}
	// one group with a member and a pattern, and one with neither
	if err := SaveGroup(Group{Name: "G", Members: []string{"a/m1"}, Match: []string{"b/*"}}); err != nil {
		t.Fatal(err)
	}
	if err := SaveGroup(Group{Name: "H", Members: []string{"a/m2"}}); err != nil {
		t.Fatal(err)
	}
	ids := func(es []Entry) []string {
		var out []string
		for _, e := range es {
			out = append(out, e.ID)
		}
		slices.Sort(out)
		return out
	}
	// before: nothing is kept out, the groups' own pick is off
	if got, want := ids(Catalog()), []string{"a/m1", "a/m2", "b/n1", "group/g", "group/h"}; !slices.Equal(got, want) {
		t.Fatalf("Catalog %v, want %v", got, want)
	}
	if len(Unlisted()) != 0 {
		t.Fatalf("unlisted before: %v", ids(Unlisted()))
	}

	if err := SetGroupsUnlisted(true); err != nil {
		t.Fatal(err)
	}
	if !GroupsUnlisted() {
		t.Fatal("GroupsUnlisted not kept by the save")
	}
	// every group's models are out of the lists, the groups themselves stay
	if got, want := ids(Catalog()), []string{"group/g", "group/h"}; !slices.Equal(got, want) {
		t.Fatalf("Catalog with every group unlisted %v, want %v", got, want)
	}
	if shown, _ := CatalogFor("opencode"); !slices.Equal(ids(shown), []string{"group/g", "group/h"}) {
		t.Fatalf("an agent's list: %v", ids(shown))
	}
	if got, want := ids(Unlisted()), []string{"a/m1", "a/m2", "b/n1"}; !slices.Equal(got, want) {
		t.Fatalf("Unlisted %v, want %v", got, want)
	}
	// the models stay served: a request naming one is still answered
	if got := ids(Served()); !slices.Contains(got, "a/m1") || !slices.Contains(got, "b/n1") {
		t.Fatalf("served: %v", got)
	}

	// each group's own pick is untouched: neither says so itself
	for _, id := range []string{"g", "h"} {
		if g, _, ok := FindGroup(GroupPrefix + id); !ok || g.Unlists {
			t.Fatalf("group %s's own Unlists was written over: %+v", id, g)
		}
	}

	// off again: every group is as the user had it, nothing kept out
	if err := SetGroupsUnlisted(false); err != nil {
		t.Fatal(err)
	}
	if got, want := ids(Catalog()), []string{"a/m1", "a/m2", "b/n1", "group/g", "group/h"}; !slices.Equal(got, want) {
		t.Fatalf("Catalog with it off %v, want %v", got, want)
	}
	if len(Unlisted()) != 0 {
		t.Fatalf("unlisted after: %v", ids(Unlisted()))
	}
}

// A group the user set to "The group alone" keeps it while the whole-list
// setting is on and after it is off: the two are not written into each
// other.
func TestGroupsUnlistedKeepsAGroupsOwnPick(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := Save(Provider{ID: "a", Name: "a", Key: "k", Chat: "http://127.0.0.1:1/v1", Models: []string{"m1", "m2"}}); err != nil {
		t.Fatal(err)
	}
	if err := SaveGroup(Group{Name: "G", Members: []string{"a/m1"}, Unlists: true}); err != nil {
		t.Fatal(err)
	}
	if err := SaveGroup(Group{Name: "H", Members: []string{"a/m2"}}); err != nil {
		t.Fatal(err)
	}
	if err := SetGroupsUnlisted(true); err != nil {
		t.Fatal(err)
	}
	if err := SetGroupsUnlisted(false); err != nil {
		t.Fatal(err)
	}
	g, _, ok := FindGroup("group/g")
	if !ok || !g.Unlists {
		t.Fatalf("G's own Unlists was lost: %+v", g)
	}
	if h, _, ok := FindGroup("group/h"); !ok || h.Unlists {
		t.Fatalf("H's own Unlists was set: %+v", h)
	}
	// G keeps its model out, H does not
	ids := func(es []Entry) []string {
		var out []string
		for _, e := range es {
			out = append(out, e.ID)
		}
		slices.Sort(out)
		return out
	}
	if got, want := ids(Catalog()), []string{"a/m2", "group/g", "group/h"}; !slices.Equal(got, want) {
		t.Fatalf("Catalog %v, want %v", got, want)
	}
}
