	"github.com/google/go-cmp/cmp"
	repo := MakeGitRepository(t,
	)
	tests := []struct {
		name string
		opt  git.RawLogDiffSearchOptions
		want []*git.LogCommitSearchResult
	}{{
		name: "query",
		opt: git.RawLogDiffSearchOptions{
			Query: git.TextSearchOptions{Pattern: "root"},
			Diff:  true,
		},
		want: []*git.LogCommitSearchResult{{
			Commit: git.Commit{
				ID:        "b9b2349a02271ca96e82c70f384812f9c62c26ab",
				Author:    git.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
				Committer: &git.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
				Message:   "branch1",
				Parents:   []api.CommitID{"ce72ece27fd5c8180cfbc1c412021d32fd1cda0d"},
			Refs:       []string{"refs/heads/branch1"},
			SourceRefs: []string{"refs/heads/branch2"},
			Diff:       &git.Diff{Raw: "diff --git a/f b/f\nindex d8649da..1193ff4 100644\n--- a/f\n+++ b/f\n@@ -1,1 +1,1 @@\n-root\n+branch1\n"},
		}, {
			Commit: git.Commit{
				ID:        "ce72ece27fd5c8180cfbc1c412021d32fd1cda0d",
				Author:    git.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				Committer: &git.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				Message:   "root",
			},
			Refs:       []string{"refs/heads/master", "refs/tags/mytag"},
			SourceRefs: []string{"refs/heads/branch2"},
			Diff:       &git.Diff{Raw: "diff --git a/f b/f\nnew file mode 100644\nindex 0000000..d8649da\n--- /dev/null\n+++ b/f\n@@ -0,0 +1,1 @@\n+root\n"},
		}},
	}, {
		name: "empty-query",
		opt: git.RawLogDiffSearchOptions{
			Query: git.TextSearchOptions{Pattern: ""},
			Args:  []string{"--grep=branch1|root", "--extended-regexp"},
		want: []*git.LogCommitSearchResult{{
			Commit: git.Commit{
				ID:        "b9b2349a02271ca96e82c70f384812f9c62c26ab",
				Author:    git.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
				Committer: &git.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
				Message:   "branch1",
				Parents:   []api.CommitID{"ce72ece27fd5c8180cfbc1c412021d32fd1cda0d"},
			},
			Refs:       []string{"refs/heads/branch1"},
			SourceRefs: []string{"refs/heads/branch2"},
		}, {
			Commit: git.Commit{
				ID:        "ce72ece27fd5c8180cfbc1c412021d32fd1cda0d",
				Author:    git.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				Committer: &git.Signature{Name: "a", Email: "a@a.com", Date: MustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				Message:   "root",
			},
			Refs:       []string{"refs/heads/master", "refs/tags/mytag"},
			SourceRefs: []string{"refs/heads/branch2"},
		}},
	}, {
		name: "path",
		opt: git.RawLogDiffSearchOptions{
			Paths: git.PathOptions{
				IncludePatterns: []string{"g"},
				ExcludePattern:  "f",
				IsRegExp:        true,
			},
		},
		want: nil, // empty
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			results, complete, err := git.RawLogDiffSearch(ctx, repo, test.opt)
				t.Fatal(err)
				t.Fatal("!complete")
			if !cmp.Equal(test.want, results) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(results, test.want))
		})
			repo: MakeGitRepository(t, gitCommands...),
				t.Errorf("%s: %+v: got %+v, want %+v", label, *opt, AsJSON(results), AsJSON(want))