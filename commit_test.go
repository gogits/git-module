package git

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCommit(t *testing.T) {
	c, err := testrepo.CatFileCommit("435ffceb7ba576c937e922766e37d4f7abdcc122")
	if err != nil {
		t.Fatal(err)
	}
	t.Run("ID", func(t *testing.T) {
		assert.Equal(t, "435ffceb7ba576c937e922766e37d4f7abdcc122", c.ID().String())
	})

	author := &Signature{
		Name:  "Jordan McCullough",
		Email: "jordan@github.com",
		When:  time.Unix(1415213395, 0),
	}
	t.Run("Author", func(t *testing.T) {
		assert.Equal(t, author.Name, c.Author().Name)
		assert.Equal(t, author.Email, c.Author().Email)
		assert.Equal(t, author.When.Unix(), c.Author().When.Unix())
	})

	t.Run("Committer", func(t *testing.T) {
		assert.Equal(t, author.Name, c.Committer().Name)
		assert.Equal(t, author.Email, c.Committer().Email)
		assert.Equal(t, author.When.Unix(), c.Committer().When.Unix())
	})

	t.Run("Message", func(t *testing.T) {
		message := `Merge pull request #35 from githubtraining/travis-yml-docker

Add special option flag for Travis Docker use case`
		assert.Equal(t, message, c.Message())
	})

	t.Run("Summary", func(t *testing.T) {
		assert.Equal(t, "Merge pull request #35 from githubtraining/travis-yml-docker", c.Summary())
	})
}

func TestCommit_Parent(t *testing.T) {
	c, err := testrepo.CatFileCommit("435ffceb7ba576c937e922766e37d4f7abdcc122")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("ParentsCount", func(t *testing.T) {
		assert.Equal(t, 2, c.ParentsCount())
	})

	t.Run("Parent", func(t *testing.T) {
		t.Run("no such parent", func(t *testing.T) {
			_, err := c.Parent(c.ParentsCount() + 1)
			assert.NotNil(t, err)
			assert.Equal(t, `revision does not exist [rev: , path: ]`, err.Error())
		})

		tests := []struct {
			n           int
			expParentID string
		}{
			{
				n:           0,
				expParentID: "a13dba1e469944772490909daa58c53ac8fa4b0d",
			},
			{
				n:           1,
				expParentID: "7c5ee6478d137417ae602140c615e33aed91887c",
			},
		}
		for _, test := range tests {
			t.Run("", func(t *testing.T) {
				p, err := c.Parent(test.n)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, test.expParentID, p.ID().String())
			})
		}
	})
}

func TestCommit_CommitByPath(t *testing.T) {
	c, err := testrepo.CatFileCommit("435ffceb7ba576c937e922766e37d4f7abdcc122")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		opt         CommitByRevisionOptions
		expCommitID string
	}{
		{
			opt: CommitByRevisionOptions{
				Path: "", // No path gets back to the commit itself
			},
			expCommitID: "435ffceb7ba576c937e922766e37d4f7abdcc122",
		},
		{
			opt: CommitByRevisionOptions{
				Path: "resources/labels.properties",
			},
			expCommitID: "755fd577edcfd9209d0ac072eed3b022cbe4d39b",
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			cc, err := c.CommitByPath(test.opt)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expCommitID, cc.ID().String())
		})
	}
}

// commitsToIDs returns a list of IDs for given commits.
func commitsToIDs(commits []*Commit) []string {
	ids := make([]string, len(commits))
	for i := range commits {
		ids[i] = commits[i].ID().String()
	}
	return ids
}

func TestCommit_CommitsByPage(t *testing.T) {
	// There are at most 5 commits can be used for pagination before this commit.
	c, err := testrepo.CatFileCommit("f5ed01959cffa4758ca0a49bf4c34b138d7eab0a")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		page         int
		size         int
		opt          CommitsByPageOptions
		expCommitIDs []string
	}{
		{
			page: 0,
			size: 2,
			expCommitIDs: []string{
				"f5ed01959cffa4758ca0a49bf4c34b138d7eab0a",
				"9cdb160ee4118035bf73c744e3bf72a1ba16484a",
			},
		},
		{
			page: 1,
			size: 2,
			expCommitIDs: []string{
				"f5ed01959cffa4758ca0a49bf4c34b138d7eab0a",
				"9cdb160ee4118035bf73c744e3bf72a1ba16484a",
			},
		},
		{
			page: 2,
			size: 2,
			expCommitIDs: []string{
				"dc64fe4ab8618a5be491a9fca46f1585585ea44e",
				"32c273781bab599b955ce7c59d92c39bedf35db0",
			},
		},
		{
			page: 3,
			size: 2,
			expCommitIDs: []string{
				"755fd577edcfd9209d0ac072eed3b022cbe4d39b",
			},
		},
		{
			page:         4,
			size:         2,
			expCommitIDs: []string{},
		},

		{
			page: 2,
			size: 2,
			opt: CommitsByPageOptions{
				Path: "src",
			},
			expCommitIDs: []string{
				"755fd577edcfd9209d0ac072eed3b022cbe4d39b",
			},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			commits, err := c.CommitsByPage(test.page, test.size, test.opt)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expCommitIDs, commitsToIDs(commits))
		})
	}
}

func TestCommit_SearchCommits(t *testing.T) {
	c, err := testrepo.CatFileCommit("2a52e96389d02209b451ae1ddf45d645b42d744c")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		pattern      string
		opt          SearchCommitsOptions
		expCommitIDs []string
	}{
		{
			pattern: "",
			expCommitIDs: []string{
				"2a52e96389d02209b451ae1ddf45d645b42d744c",
				"57d0bf61e57cdacb309ebd1075257c6bd7e1da81",
				"cb2d322bee073327e058143329d200024bd6b4c6",
				"818f033c4ae7f26b2b29e904942fa79a5ccaadd0",
				"369adba006a1bbf25e957a8622d2b919c994d035",
				"2956e1d20897bf6ed509f6429d7f64bc4823fe33",
				"333fd9bc94084c3e07e092e2bc9c22bab4476439",
				"f5ed01959cffa4758ca0a49bf4c34b138d7eab0a",
				"9cdb160ee4118035bf73c744e3bf72a1ba16484a",
				"dc64fe4ab8618a5be491a9fca46f1585585ea44e",
				"32c273781bab599b955ce7c59d92c39bedf35db0",
				"755fd577edcfd9209d0ac072eed3b022cbe4d39b",
			},
		},
		{
			pattern: "",
			opt: SearchCommitsOptions{
				MaxCount: 3,
			},
			expCommitIDs: []string{
				"2a52e96389d02209b451ae1ddf45d645b42d744c",
				"57d0bf61e57cdacb309ebd1075257c6bd7e1da81",
				"cb2d322bee073327e058143329d200024bd6b4c6",
			},
		},

		{
			pattern: "feature",
			expCommitIDs: []string{
				"2a52e96389d02209b451ae1ddf45d645b42d744c",
				"cb2d322bee073327e058143329d200024bd6b4c6",
			},
		},
		{
			pattern: "feature",
			opt: SearchCommitsOptions{
				MaxCount: 1,
			},
			expCommitIDs: []string{
				"2a52e96389d02209b451ae1ddf45d645b42d744c",
			},
		},

		{
			pattern: "add.*",
			opt: SearchCommitsOptions{
				Path: "src",
			},
			expCommitIDs: []string{
				"cb2d322bee073327e058143329d200024bd6b4c6",
				"818f033c4ae7f26b2b29e904942fa79a5ccaadd0",
				"333fd9bc94084c3e07e092e2bc9c22bab4476439",
				"32c273781bab599b955ce7c59d92c39bedf35db0",
				"755fd577edcfd9209d0ac072eed3b022cbe4d39b",
			},
		},
		{
			pattern: "add.*",
			opt: SearchCommitsOptions{
				MaxCount: 2,
				Path:     "src",
			},
			expCommitIDs: []string{
				"cb2d322bee073327e058143329d200024bd6b4c6",
				"818f033c4ae7f26b2b29e904942fa79a5ccaadd0",
			},
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			commits, err := c.SearchCommits(test.pattern, test.opt)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expCommitIDs, commitsToIDs(commits))
		})
	}
}
