package url_backend

import (
	"testing"
)

func TestIdToUrl1(t *testing.T) {
	url, err := IdToUrl(0)
	if err != nil {
		t.Error("For", 0, "got exception")
	}

	if url != "aaaaaaa" {
		t.Error("For", 0, "expected", "aaaaaaa", "got", url)
	}
}

func TestIdToUrl2(t *testing.T) {
	url, err := IdToUrl(62)
	if err != nil {
		t.Error("For", 62, "got exception")
	}

	if url != "abaaaaa" {
		t.Error("For", 62, "expected", "abaaaaa", "got", url)
	}
}

func TestIdToUrl3(t *testing.T) {
	url, err := IdToUrl(1 + 62 + 62*62)
	if err != nil {
		t.Error("For", 1+62+62*62, "got exception")
	}

	if url != "bbbaaaa" {
		t.Error("For", 1+62+62*62, "expected", "bbbaaaa", "got", url)
	}
}

func TestUrlToId1(t *testing.T) {
	id, err := UrlToId("aaaaaaa")
	if err != nil {
		t.Error("For", "aaaaaaa", "got exception")
	}

	if id != 0 {
		t.Error("For", "aaaaaaa", "expected", 0, "got", id)
	}
}

func TestUrlToId2(t *testing.T) {
	id, err := UrlToId("abaaaaa")
	if err != nil {
		t.Error("For", "aaaaaaa", "got exception")
	}

	if id != 62 {
		t.Error("For", "abaaaaa", "expected", 62, "got", id)
	}
}

func TestUrlToId3(t *testing.T) {
	id, err := UrlToId("bbbaaaa")
	if err != nil {
		t.Error("For", "bbbaaaa", "got exception")
	}

	if id != 1+62+62*62 {
		t.Error("For", "bbbaaaa", "expected", 1+62+62*62, "got", id)
	}
}

func TestComposition1(t *testing.T) {
	var id int64 = 214232341
	url, err := IdToUrl(id)
	if err != nil {
		t.Error("Got exception")
	}

	idRet, err := UrlToId(url)

	if idRet != id {
		t.Error("For", id, "expected", id, "got", idRet)
	}
}

func TestComposition2(t *testing.T) {
	var id int64 = 57731386987
	url, err := IdToUrl(id)
	if err != nil {
		t.Error("Got exception")
	}

	idRet, err := UrlToId(url)
	if err != nil {
		t.Error("Got exception")
	}

	if idRet != id {
		t.Error("For", id, "expected", id, "got", idRet)
	}
}

func TestComposition3(t *testing.T) {
	var id int64 = 314234321234
	url, err := IdToUrl(id)
	if err != nil {
		t.Error("Got exception")
	}

	idRet, err := UrlToId(url)
	if err != nil {
		t.Error("Got exception")
	}

	if idRet != id {
		t.Error("For", id, "expected", id, "got", idRet)
	}
}
