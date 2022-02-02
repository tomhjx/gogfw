package gogfw

import (
	"testing"
)

var types = map[string]bool{
	ITEM_TYPE_DOMAIN:        true,
	ITEM_TYPE_DOMAIN_SUFFIX: true,
	ITEM_TYPE_IP:            true,
}

func TestOffline(t *testing.T) {
	var (
		hd    *Handle
		err   error
		items []*Item
	)
	hd, err = OpenOffline("/work/resources/gfwlist.txt")
	if err != nil {
		t.Error(err)
	}
	items, err = hd.ReadItems()
	if err != nil {
		t.Error(err)
	}

	for _, v := range items {
		// t.Log(v.Type, v.Value)
		if !types[v.Type] {
			t.Errorf("[%s] %s is error.", v.Type, v.Value)
		}
	}
}

func TestOnline(t *testing.T) {
	var (
		hd    *Handle
		err   error
		items []*Item
	)
	hd, err = OpenOnline("https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt")
	if err != nil {
		t.Error(err)
	}
	items, err = hd.ReadItems()
	if err != nil {
		t.Error(err)
	}
	for _, v := range items {
		// t.Log(v.Type, v.Value)
		if !types[v.Type] {
			t.Errorf("[%s] %s is error.", v.Type, v.Value)
		}
	}
}

func TestParseItem(t *testing.T) {

	// Adblock Plus filters
	ts := []string{
		"!--||adorama.com",
		"*share*",
		"share.dmhy.org",
		"video.aol.ca/video-detail",
		"wikilivres.info/wiki/%E9%9B%B6%E5%85%AB%E5%AE%AA%E7%AB%A0",
		"lists.w3.org/archives/public",
		".casinobellini.com",
		".google.*/falun",
		"|facebook*",
		"|http://85.17.73.31/",
		"|http://85.17.73.31:8080/",
		"|http://85.17.73.31:8080/abc/efg#",
		"1.2.3.4",
		"50.7.31.230:8898",
		"|http://www.dmm.com/netgame",
		"|https://raw.githubusercontent.com/programthink/zhao",
		"|http://imgmega.com/*.gif.html",
		"|http://*.pimg.tw/",
		"|https://*.s3.amazonaws.com",
		"||afreecatv.com",
		"||cdn*.i-scmp.com",
		"||xn--4gq171p.com",
		"||xn--ngstr-lra8j.com",
		"@@||cn.noxinfluencer.com",
	}
	hd := Handle{}

	for _, v := range ts {
		item, err := hd.ParseItem(v)
		if err != nil {
			t.Log(err)
			continue
		}
		if item == nil {
			t.Logf("ignore: %s", v)
			continue
		}
		t.Logf("item: %s, %s", item.Type, item.Value)
	}

}
