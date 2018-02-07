package main

import "testing"

func TestMain(m *testing.M) {
	meta1 := &ImageMeta{
		Title:              "เรื่องสยองของอดัม byเรื่องสยองของแอนนี่ (อ่านฟรี)",
		Name:               "ANNYSTORY ANNYSTORY ANNYSTORY ANNYSTORY ANNYSTORY ANNYSTORY",
		BackgroundImageURL: "",
		ProfileImageURL:    "https://static.jamplay.world/author/5a5dc2beffa696001255358a/f75473a2-1f60-4d0e-ad88-d7fb6193b8b4.blob.jpg",
		ThumbnailImageURL:  "https://static.jamplay.world/book/5a65aedc14a5f6000f2b291d/4b38afba-2968-4ab6-8250-d14590d14959.blob.jpg",
		FileName:           "test-render-1.png",
	}
	DrawImage(meta1)
}
