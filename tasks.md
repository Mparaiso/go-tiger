go-tiger tasks
==============

x = done 

priorities : H = high , L = low

funcs

- [x] add funcs.MakeEvery
- [x] add funcs.MakeSome
- [x] add funcs.MakeFilter
- [L] add funcs.MakeLastIndexOf
- [x] add funcs.MakeFind
- [L] add funcs.MakeReverse
- [H] add funcs.MakeSort
- [x] add funcs.MakeForEach
- [x] add funcs.MakeInclude
- [L] add funcs.MakeDifference
- [L] add funcs.MakeUnion
- [L] add funcs.MakeXor
- [x] add funcs.MakeGroupBy https://lodash.com/docs/4.17.2#groupBy 
- [L] add funcs.MakePartition https://lodash.com/docs/4.17.2#partition
- [L] add funcs.MakeShuffle
- [x] add funcs.KeyBy https://lodash.com/docs/4.17.2#keyBy `func(collection []T,keyProvider func(element T)K)map[K]T`
- [x] add funcs.GetKeys
- [x] add funcs.getValues

mongo 

- [x] add branch mongo : complete DocumentManager.resolveAllRelations
- [H] add support for resolveAllRelations/referenceMany/MappedBy 
- [H] add support for inversedBy annotation
- [L] add support for order in mapping
- [L] add support for specific criteria in mapping
- [L] add support for limit in mapping
- [H] fix unity of work , make sure a recursive cascade doesn't lead to a stack overflow
