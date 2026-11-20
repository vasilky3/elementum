package bittorrent

import "testing"

func TestMatchSeasonFolderEpisode(t *testing.T) {
	cases := []struct {
		path    string
		season  int
		episode int
		want    bool
	}{
		{`Season_01\17. Забытая сделка.mkv`, 1, 17, true},
		{`Season_01/17. Забытая сделка.mkv`, 1, 17, true},
		{`Season_01\18. Когда зло становится добром.mkv`, 1, 18, true},
		{`Season_01\18. Когда зло становится добром.mkv`, 1, 17, false},
		{`Season_02\01. Испытание ведьм.mkv`, 2, 1, true},
		{`Season_02\01. Испытание ведьм.mkv`, 1, 1, false},
		{`Season_02\05. Она — мужчина!.mkv`, 2, 5, true},
		{`Season_02/17. Другая.mkv`, 1, 17, false},
		{`Season_02/17. Другая.mkv`, 2, 17, true},
		{`Season_11/01. Title.mkv`, 1, 1, false},
		{`Season_11/01. Title.mkv`, 11, 1, true},
		{`Season_01/11. Title.mkv`, 1, 1, false},
		{`Season_01/11. Title.mkv`, 1, 11, true},
		{`Season_01/01. Title.mkv`, 1, 1, true},
		{`Season_01/07. Title.mkv`, 1, 17, false},
		{`Season_01/17. Title.mkv`, 1, 7, false},
		{`Show/Season_01/17. Title.mkv`, 1, 17, true},
		{`Show.1998/Season_02/01. Title.mkv`, 2, 1, true},
		{`Show.1998/Season_02/01. Title.mkv`, 1, 1, false},
		{`Season 01/17 - Title.mkv`, 1, 17, true},
		{`S01/17.mkv`, 1, 17, true},
		{`Staffel_02/05. Title.mkv`, 2, 5, true},
		{`Сезон_01/17. Title.mkv`, 1, 17, true},
		{`01/17. Title.mkv`, 1, 17, true},
		{`17. Забытая сделка.mkv`, 1, 17, false},
		{`Season_01/17. Title.mkv`, 2, 17, false},
	}

	for _, tc := range cases {
		if got := matchSeasonFolderEpisode(tc.path, tc.season, tc.episode); got != tc.want {
			t.Errorf("matchSeasonFolderEpisode(%q, %d, %d) = %v, want %v", tc.path, tc.season, tc.episode, got, tc.want)
		}
	}
}

func TestMatchEpisodeFilenameSeasonFolders(t *testing.T) {
	choices := []*CandidateFile{
		{Filename: "17. Забытая сделка.mkv", Path: `Season_01\17. Забытая сделка.mkv`},
		{Filename: "18. Когда зло становится добром.mkv", Path: `Season_01\18. Когда зло становится добром.mkv`},
		{Filename: "01. Испытание ведьм.mkv", Path: `Season_02\01. Испытание ведьм.mkv`},
		{Filename: "05. Она — мужчина!.mkv", Path: `Season_02\05. Она — мужчина!.mkv`},
		{Filename: "17. Другая.mkv", Path: `Season_02\17. Другая.mkv`},
		{Filename: "01. Title.mkv", Path: `Season_11\01. Title.mkv`},
	}

	index, found := MatchEpisodeFilename(1, 17, false, 1, nil, nil, nil, choices)
	if found != 1 || index != 0 {
		t.Fatalf("S01E17: index=%d found=%d, want index 0 found 1", index, found)
	}

	index, found = MatchEpisodeFilename(2, 1, false, 2, nil, nil, nil, choices)
	if found != 1 || index != 2 {
		t.Fatalf("S02E01: index=%d found=%d, want index 2 found 1", index, found)
	}

	index, found = MatchEpisodeFilename(2, 5, false, 2, nil, nil, nil, choices)
	if found != 1 || index != 3 {
		t.Fatalf("S02E05: index=%d found=%d, want index 3 found 1", index, found)
	}

	index, found = MatchEpisodeFilename(11, 1, false, 11, nil, nil, nil, choices)
	if found != 1 || index != 5 {
		t.Fatalf("S11E01: index=%d found=%d, want index 5 found 1", index, found)
	}
}
