insidetree
==========

This datastructure is specialized in indexing S2 cells to perform stab queries (Point in Polygons).

For explanations about s2 and the tree used in this package look at this [great blogpost](https://blog.zen.ly/geospatial-indexing-on-hilbert-curves-2379b929addc)

It's an alternative of using an interval tree like [in this other project](https://github.com/akhenakh/regionagogo) exposed [in this blogpost](https://blog.nobugware.com/post/2016/geo_db_s2_region_polygon/).


```go
tree := NewTree()

// index you cells using Index()
tree.Index(cell, "Red building")

// test for cell of any levels 
insideCell := s2.CellIDFromLatLng(s2.LatLngFromDegrees(46.83808, -71.28046))
results := tree.Stab(insideCell)
// results[0] == "Red Building"
```