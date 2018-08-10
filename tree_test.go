package insidetree

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	"github.com/golang/geo/s2"
	"github.com/stretchr/testify/require"
	geom "github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

var result []interface{}

func TestIndex(t *testing.T) {
	tree := NewTree()

	ll := s2.LatLngFromDegrees(48.846043, 2.336943)
	cid := s2.CellIDFromLatLng(ll)
	cid = cid.Parent(15)

	all := s2.LatLngFromDegrees(48.843366, 2.334117)
	acid := s2.CellIDFromLatLng(all)
	acid = acid.Parent(15)

	t.Log(cid, acid, cid.Parent(14), acid.Parent(14))

	tree.Index(cid, 41)
	tree.Index(acid, 42)

	// stabing at the same level
	res := tree.Stab(cid)
	require.Len(t, res, 1)
	require.Equal(t, res[0], 41)

	// mask at the same level
	res = tree.Mask(cid)
	require.Len(t, res, 1)
	require.Equal(t, res[0], 41)

	// stabing at level 30
	cid = s2.CellIDFromLatLng(ll)
	res = tree.Stab(cid)
	require.Len(t, res, 1)
	require.Equal(t, res[0], 41)

	// stabing at higher level 10 (normally not what you want)
	cid = s2.CellIDFromLatLng(ll)
	cid = cid.Parent(10)
	res = tree.Stab(cid)
	require.Len(t, res, 0)

	// mask at higher level 10
	res = tree.Mask(cid)
	require.Len(t, res, 2)

	ll = s2.LatLngFromDegrees(48.844586, 2.333863)
	cid = s2.CellIDFromLatLng(ll)
	cid = cid.Parent(14)

	// mask at higher level 10
	// would return everything inserted under
	res = tree.Mask(cid)
	require.Len(t, res, 2)
}

func TestGeoJSONPolygons(t *testing.T) {
	tree := NewTree()

	err := prepareData(tree)
	require.NoError(t, err)

	outsideCell := s2.CellIDFromLatLng(s2.LatLngFromDegrees(46.83643, -71.27638))
	res := tree.Stab(outsideCell)
	require.Len(t, res, 0)

	insideCell := s2.CellIDFromLatLng(s2.LatLngFromDegrees(46.83704, -71.27751))
	res = tree.Stab(insideCell)
	require.Len(t, res, 1)
	require.Equal(t, res[0], "27719949")

	insideCell = s2.CellIDFromLatLng(s2.LatLngFromDegrees(46.83808, -71.28046))
	res = tree.Stab(insideCell)
	require.Len(t, res, 1)
	require.Equal(t, res[0], "128447237")

	insideCell = s2.CellIDFromLatLng(s2.LatLngFromDegrees(46.79900, -71.23723))
	res = tree.Stab(insideCell)
	require.Len(t, res, 1)
	require.Equal(t, res[0], "103864989")

	// this point is in 2 shapes
	insideCell = s2.CellIDFromLatLng(s2.LatLngFromDegrees(46.79884, -71.23748))
	res = tree.Stab(insideCell)
	require.Len(t, res, 2)

	start := time.Now()
	for i := 0; i < 100000; i++ {
		res = tree.Stab(insideCell)
	}

	t.Log(time.Since(start), res)
}

func BenchmarkLookup(b *testing.B) {
	tree := NewTree()
	err := prepareData(tree)
	require.NoError(b, err)
	insideCell := s2.CellIDFromLatLng(s2.LatLngFromDegrees(46.83808, -71.28046))

	b.ResetTimer()

	var res []interface{}
	for n := 0; n < b.N; n++ {
		r := tree.Stab(insideCell)
		res = r
	}

	result = res
}

func prepareData(tree *Tree) error {
	f, err := ioutil.ReadFile("testdata/buildins_quebec.geojson")
	if err != nil {
		return err
	}
	js := geojson.FeatureCollection{}

	err = json.Unmarshal(f, &js)
	if err != nil {
		return err
	}

	coverer := s2.RegionCoverer{MaxLevel: 22, MaxCells: 100}
	for _, f := range js.Features {
		switch g := f.Geometry.(type) {
		case *geom.MultiPolygon:
			rawCoords := g.FlatCoords()
			coords := g.FlatCoords()[0 : len(rawCoords)-2]
			l := LoopFromCoordinates(coords)
			if l == nil || !l.IsValid() || !l.HasInterior() || l.IsFull() || l.ContainsOrigin() {
				continue
			}
			cu := coverer.Covering(l)
			for _, c := range cu {
				if c.Level() < 1 {
					continue
				}
				tree.Index(c, f.Properties["osm_way_id"])
			}
		}
	}
	return nil
}

// LoopFromCoordinates creates a LoopFence from a list of lng lat
func LoopFromCoordinates(c []float64) *s2.Loop {
	if len(c)%2 != 0 || len(c) <= 2*3 {
		return nil
	}
	points := make([]s2.Point, len(c)/2)

	for i := 0; i < len(c); i += 2 {
		points[i/2] = s2.PointFromLatLng(s2.LatLngFromDegrees(c[i+1], c[i]))
	}

	if s2.RobustSign(points[0], points[1], points[2]) != s2.CounterClockwise {
		// reversing the slice
		for i := len(points)/2 - 1; i >= 0; i-- {
			opp := len(points) - 1 - i
			points[i], points[opp] = points[opp], points[i]
		}
	}

	loop := s2.LoopFromPoints(points)
	return loop
}
