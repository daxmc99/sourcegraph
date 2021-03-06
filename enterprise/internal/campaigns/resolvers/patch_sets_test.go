package resolvers

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func TestPatchSetResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)
	rcache.SetupForTest(t)

	// For testing purposes they all share the same rev, across repos
	testingRev := api.CommitID("24f7ca7c1190835519e261d7eefa09df55ceea4f")

	mockBackendCommits(t, testingRev)

	repoupdater.MockRepoLookup = func(args protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
		return &protocol.RepoLookupResult{
			Repo: &protocol.RepoInfo{Name: args.Repo},
		}, nil
	}
	defer func() { repoupdater.MockRepoLookup = nil }()

	reposStore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})

	var rs []*repos.Repo
	for i := 0; i < 3; i++ {
		repo := newGitHubTestRepo(fmt.Sprintf("github.com/sourcegraph/sourcegraph-%d", i), i)
		err := reposStore.UpsertRepos(ctx, repo)
		if err != nil {
			t.Fatal(err)
		}
		rs = append(rs, repo)
	}

	store := ee.NewStore(dbconn.Global)

	userID := insertTestUser(t, dbconn.Global, "patch-set-resolver", true)
	patchSet := &campaigns.PatchSet{UserID: userID}
	err := store.CreatePatchSet(ctx, patchSet)
	if err != nil {
		t.Fatal(err)
	}

	var (
		testDiffStatAdded   int32 = 0
		testDiffStatDeleted int32 = 0
		testDiffStatChanged int32 = 2
	)

	var patches []*campaigns.Patch
	for _, repo := range rs {
		patch := &campaigns.Patch{
			PatchSetID:      patchSet.ID,
			RepoID:          repo.ID,
			Rev:             testingRev,
			BaseRef:         "master",
			Diff:            testDiff,
			DiffStatAdded:   &testDiffStatAdded,
			DiffStatDeleted: &testDiffStatDeleted,
			DiffStatChanged: &testDiffStatChanged,
		}

		err := store.CreatePatch(ctx, patch)
		if err != nil {
			t.Fatal(err)
		}
		patches = append(patches, patch)
	}

	sr := &Resolver{store: store}
	s, err := graphqlbackend.NewSchema(sr, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	var response struct{ Node apitest.PatchSet }
	apitest.MustExec(ctx, t, s, nil, &response, fmt.Sprintf(`
        query {
          node(id: %q) {
            ... on PatchSet {
              id
              diffStat {
                added
                deleted
                changed
              }
              patches(first: %d) {
                nodes {
				  ... on HiddenPatch {
				    id
				  }
				  ... on Patch {
                    repository {
                      name
                    }
                    diff {
                      fileDiffs(first: %d, after: %s) {
                        rawDiff
                        diffStat {
                          added
                          deleted
                          changed
                        }
                        pageInfo {
                          endCursor
                          hasNextPage
                        }
                        nodes {
                          oldPath
                          newPath
                          hunks {
                            body
                            section
                            newRange { startLine, lines }
                            oldRange { startLine, lines }
                            oldNoNewlineAt
                          }
                          stat {
                            added
                            deleted
                            changed
                          }
                          oldFile {
                            name
                            externalURLs {
                              serviceType
                              url
                            }
                          }
                        }
                      }
                    }
				  }
                }
              }
            }
          }
        }
		`, marshalPatchSetID(patchSet.ID), len(patches), 10000, "null"))

	if have, want := len(response.Node.Patches.Nodes), len(patches); have != want {
		t.Fatalf("have %d patches, want %d", have, want)
	}

	// Each patch has testDiff as diff, each with 2 lines changed
	if have, want := response.Node.DiffStat.Changed, int32(len(patches)*2); have != want {
		t.Fatalf("wrong PatchSet.DiffStat.Changed %d, want=%d", have, want)
	}

	for i, patch := range response.Node.Patches.Nodes {
		if have, want := patch.Repository.Name, rs[i].Name; have != want {
			t.Fatalf("wrong Repository Name %q. want=%q", have, want)
		}

		if have, want := patch.Diff.FileDiffs.RawDiff, testDiff; have != want {
			t.Fatalf("wrong RawDiff. diff=%s", cmp.Diff(have, want))
		}

		if have, want := patch.Diff.FileDiffs.DiffStat.Changed, int32(2); have != want {
			t.Fatalf("wrong DiffStat.Changed %d, want=%d", have, want)
		}

		haveFileDiffs := patch.Diff.FileDiffs
		if !reflect.DeepEqual(haveFileDiffs, testDiffGraphQL) {
			t.Fatal(cmp.Diff(haveFileDiffs, testDiffGraphQL))
		}
	}
}

func TestPatchSetsFileDiffs(t *testing.T) {
	ctx := context.Background()

	const testDiffFiles = 3
	const testOldFile = `First
Second
Line 1
Line 2
Line 3
Line 4
Line 5
Line 6
Line 7
Line 8
Line 9
Line 10
`

	const testDiff = `diff --git INSTALL.md INSTALL.md
index e5af166..d44c3fc 100644
--- INSTALL.md
+++ INSTALL.md
@@ -3,10 +3,10 @@
 Line 1
 Line 2
 Line 3
-Line 4
+This is cool: Line 4
 Line 5
 Line 6
-Line 7
-Line 8
+Another Line 7
+Foobar Line 8
 Line 9
 Line 10
diff --git JOKES.md JOKES.md
index ea80abf..1b86505 100644
--- JOKES.md
+++ JOKES.md
@@ -4,10 +4,10 @@ Joke #1
 Joke #2
 Joke #3
 Joke #4
-Joke #5
+This is not funny: Joke #5
 Joke #6
-Joke #7
+This one is good: Joke #7
 Joke #8
-Joke #9
+Waffle: Joke #9
 Joke #10
 Joke #11
diff --git README.md README.md
index 9bd8209..d2acfa9 100644
--- README.md
+++ README.md
@@ -1,12 +1,13 @@
 # README

-Line 1
+Foobar Line 1
 Line 2
 Line 3
 Line 4
 Line 5
-Line 6
+Barfoo Line 6
 Line 7
 Line 8
 Line 9
 Line 10
+Another line
`

	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := func() time.Time {
		return now.UTC().Truncate(time.Microsecond)
	}

	wantBaseRef := "refs/heads/master"
	wantHeadRevision := api.CommitID("b69072d5f687b31b9f6ae3ceafdc24c259c4b9ec")
	mockBackendCommits(t, api.CommitID(wantBaseRef), wantHeadRevision)

	repo := &types.Repo{ID: api.RepoID(1), Name: "github.com/sourcegraph/sourcegraph"}

	patch := &patchResolver{
		store: ee.NewStoreWithClock(dbconn.Global, clock),
		patch: &campaigns.Patch{
			RepoID:  repo.ID,
			Rev:     api.CommitID(wantHeadRevision),
			BaseRef: wantBaseRef,
			Diff:    testDiff,
		},
		preloadedRepo: repo,
	}

	t.Run("Pagination", func(t *testing.T) {
		endCursors := []string{"1", "2"}
		totalCount := int32(testDiffFiles)

		tests := []struct {
			first int32
			after string

			wantNodeCount int

			wantTotalCount *int32

			wantHasNextPage bool
			wantEndCursor   *string
		}{
			{
				first:           1,
				after:           "",
				wantNodeCount:   1,
				wantHasNextPage: true,
				wantEndCursor:   &endCursors[0],
				wantTotalCount:  nil,
			},
			{
				first:           1,
				after:           endCursors[0],
				wantNodeCount:   1,
				wantHasNextPage: true,
				wantEndCursor:   &endCursors[1],
				wantTotalCount:  nil,
			},
			{
				first:           1,
				after:           endCursors[1],
				wantNodeCount:   1,
				wantHasNextPage: false,
				wantEndCursor:   nil,
				wantTotalCount:  &totalCount,
			},
			{
				first:           testDiffFiles + 1,
				after:           "",
				wantNodeCount:   testDiffFiles,
				wantHasNextPage: false,
				wantEndCursor:   nil,
				wantTotalCount:  &totalCount,
			},
		}

		for _, tc := range tests {
			args := &graphqlbackend.FileDiffsConnectionArgs{First: &tc.first}
			if tc.after != "" {
				args.After = &tc.after
			}

			conn, err := patch.FileDiffs(ctx, args)
			if err != nil {
				t.Fatal(err)
			}

			nodes, err := conn.Nodes(ctx)
			if err != nil {
				t.Fatal(err)
			}

			if len(nodes) != tc.wantNodeCount {
				t.Fatalf("wrong length of nodes. want=%d, have=%d", tc.wantNodeCount, len(nodes))
			}

			pageInfo, err := conn.PageInfo(ctx)
			if err != nil {
				t.Fatal(err)
			}

			if pageInfo.HasNextPage() != tc.wantHasNextPage {
				t.Fatalf("pageInfo HasNextPage wrong. want=%t, have=%t", tc.wantHasNextPage, pageInfo.HasNextPage())
			}

			if diff := cmp.Diff(tc.wantEndCursor, pageInfo.EndCursor()); diff != "" {
				t.Fatalf("(-want +got):\n%s", diff)
			}

			totalCount, err := conn.TotalCount(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.wantTotalCount, totalCount); diff != "" {
				t.Fatalf("wrong totalCount: %s", diff)
			}
		}
	})

	t.Run("NewFile resolver", func(t *testing.T) {
		fileDiffConnection, err := patch.FileDiffs(ctx, &graphqlbackend.FileDiffsConnectionArgs{})
		if err != nil {
			t.Fatal(err)
		}
		fileDiffs, err := fileDiffConnection.Nodes(ctx)
		if err != nil {
			t.Fatal(err)
		}

		if have, want := len(fileDiffs), testDiffFiles; have != want {
			t.Fatalf("invalid len(FileDiffs.Nodes). want=%d have=%d", want, len(fileDiffs))
		}
		fileDiff := fileDiffs[0]

		git.Mocks.ReadFile = func(commit api.CommitID, name string) ([]byte, error) {
			if name != "INSTALL.md" {
				t.Fatalf("ReadFile received call for wrong file: %s", name)
			}

			return []byte(testOldFile), nil
		}
		defer func() { git.Mocks.ReadFile = nil }()

		newFile := fileDiff.NewFile()
		if newFile == nil {
			t.Fatal("NewFile is null")
		}

		wantNewFileContent := `First
Second
Line 1
Line 2
Line 3
This is cool: Line 4
Line 5
Line 6
Another Line 7
Foobar Line 8
Line 9
Line 10`

		haveContent, err := newFile.Content(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if haveContent != wantNewFileContent {
			t.Fatalf("wrong file content. want=%q have=%q", wantNewFileContent, haveContent)
		}
	})
}

func TestApplyPatch(t *testing.T) {
	tests := []struct {
		file          string
		patch         string
		origStartLine int32
		wantFile      string
	}{
		{
			file: `1 some
2
3
4
5
6
7 super awesome
8
9
10
11
12
13
14 file
15
16
17
18 oh yes`,
			patch: ` 4
 5
 6
-7 super awesome
+7 super mega awesome
 8
 9
 10
`,
			origStartLine: 4,
			wantFile: `1 some
2
3
4
5
6
7 super mega awesome
8
9
10
11
12
13
14 file
15
16
17
18 oh yes`,
		},
	}

	for _, tc := range tests {
		have := applyPatch(tc.file, &diff.FileDiff{Hunks: []*diff.Hunk{{OrigStartLine: tc.origStartLine, Body: []byte(tc.patch)}}})
		if have != tc.wantFile {
			t.Fatalf("wrong patched file content %q, want=%q", have, tc.wantFile)
		}
	}
}
