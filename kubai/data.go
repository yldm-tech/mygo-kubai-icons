package kubai

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"sync"
)

// The generated data. A family is a base icon and its badged and slashed composites, packed into one deflated, base64-encoded JSON blob (gen/generate.mjs): icon name to icon data.

// family decodes its blob the first time one of its icons is asked for. A family no icon function references is left out of the executable by the linker.
type family struct {
	once  sync.Once
	icons map[string]*Icon
}

func (f *family) icon(data, name string) *Icon {
	f.once.Do(func() { f.icons = decodeFamily(data) })
	return f.icons[name]
}

type iconData struct {
	Tone      string     `json:"tone"`
	Tap       bool       `json:"tap"`
	Templates []template `json:"templates"`
	Tracks    []track    `json:"tracks"`
}

// A template is markup with holes: Text[0], then for each i the number in slot Refs[i] followed by Text[i+1]. Main markup and shadow markup are separated by sep.
type template struct {
	Text []string `json:"text"`
	Refs []int    `json:"refs"`
}

// A track holds the keyframes of the hover motion for one tap state, playing forward or, when Out, back to rest. Each key is [progress, template, slot values...].
type track struct {
	Active bool        `json:"active"`
	Out    bool        `json:"out"`
	Keys   [][]float64 `json:"keys"`
}

const sep = "\x01"

// decodeFamily panics on data in error: the data is generated and tested, so an error is a bug in this module, not in the program using it.
func decodeFamily(data string) map[string]*Icon {
	raw, err := io.ReadAll(flate.NewReader(base64.NewDecoder(base64.StdEncoding, strings.NewReader(strings.ReplaceAll(data, "\n", "")))))
	if err != nil {
		panic("kubai: corrupt icon data: " + err.Error())
	}
	var m map[string]iconData
	if err := json.NewDecoder(bytes.NewReader(raw)).Decode(&m); err != nil {
		panic("kubai: corrupt icon data: " + err.Error())
	}
	out := make(map[string]*Icon, len(m))
	for name, d := range m {
		out[name] = &Icon{name: name, tone: d.Tone, tap: d.Tap, templates: d.Templates, tracks: d.Tracks}
	}
	return out
}

// pose is one moment of the motion: a template and its slot values.
type pose struct {
	t     int
	slots []float64
}

// at returns the pose of the track at progress v, interpolating between the keyframes around it when they share a template, and taking the nearer one when they do not.
func (tr *track) at(v float64) pose {
	keys := tr.Keys
	v = min(max(v, 0), 1)
	i := 1
	for i < len(keys)-1 && keys[i][0] < v {
		i++
	}
	a, b := keys[i-1], keys[i]
	span := b[0] - a[0]
	k := 0.0
	if span > 0 {
		k = (v - a[0]) / span
	}
	if a[1] != b[1] {
		if k < 0.5 {
			return pose{t: int(a[1]), slots: a[2:]}
		}
		return pose{t: int(b[1]), slots: b[2:]}
	}
	return pose{t: int(a[1]), slots: lerp(a[2:], b[2:], k)}
}

func lerp(a, b []float64, k float64) []float64 {
	out := make([]float64, len(a))
	for i := range a {
		out[i] = a[i] + (b[i]-a[i])*k
	}
	return out
}

// render writes the markup of a pose: the icon's layers, and the silhouette its shadow is cast from (empty when nothing casts one).
func (t *template) render(slots []float64) (main, shadow string) {
	var sb strings.Builder
	sb.WriteString(t.Text[0])
	var buf [24]byte
	for i, r := range t.Refs {
		sb.Write(fmtNum(buf[:0], slots[r]))
		sb.WriteString(t.Text[i+1])
	}
	main, shadow, _ = strings.Cut(sb.String(), sep)
	return main, shadow
}

// fmtNum writes n with at most three decimals, as upstream rounds its geometry.
func fmtNum(b []byte, n float64) []byte {
	b = strconv.AppendFloat(b, float64(int64(n*1000+copysignHalf(n)))/1000, 'f', -1, 64)
	return b
}

func copysignHalf(n float64) float64 {
	if n < 0 {
		return -0.5
	}
	return 0.5
}
