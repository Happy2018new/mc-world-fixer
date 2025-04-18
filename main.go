package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"sync"

	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/neomega/chunks/chunk"
	"github.com/OmineDev/neomega-core/neomega/chunks/define"
	"github.com/OmineDev/neomega-core/utils/game_saves/bedrock_level"
	bedrock_level_operation "github.com/OmineDev/neomega-core/utils/game_saves/bedrock_level/operation"
	bedrock_level_operation_define "github.com/OmineDev/neomega-core/utils/game_saves/bedrock_level/operation/define"
	bedrock_level_operation_marshal "github.com/OmineDev/neomega-core/utils/game_saves/bedrock_level/operation/marshal"
	bedrock_level_provider "github.com/OmineDev/neomega-core/utils/game_saves/bedrock_level/provider"
	"github.com/df-mc/goleveldb/leveldb/opt"
	"github.com/pterm/pterm"
)

type SubChunkAndHash struct {
	SubChunk    *chunk.SubChunk
	InternalIdx int
	Hash        uint64
}

func QuickGetSubChunkIndex(singleSubChunk []byte, r define.Range) (int, error) {
	buf := bytes.NewBuffer(singleSubChunk)

	ver, err := buf.ReadByte()
	if err != nil {
		return -1, fmt.Errorf("error reading version: %w", err)
	}

	switch ver {
	default:
		return -1, fmt.Errorf("unknown sub chunk version %v: can't decode", ver)
	case 1:
		return -1, fmt.Errorf("should not happened")
		// NOP
	case 8, 9:
		// Version 8 allows up to 256 layers for one sub chunk.
		_, err = buf.ReadByte()
		if err != nil {
			return -1, fmt.Errorf("error reading storage count: %w", err)
		}

		if ver == 9 {
			uIndex, err := buf.ReadByte()
			if err != nil {
				return -1, fmt.Errorf("error reading subchunk index: %w", err)
			}
			// The index as written here isn't the actual index of the subchunk within the chunk. Rather, it is the Y
			// value of the subchunk. This means that we need to translate it to an index.
			return int(int8(uIndex)), nil
		}

		panic("??????")
	}
}

func LoadSubChunk(db *bedrock_level_provider.Provider, dm define.Dimension, position protocol.SubChunkPos, keyIdx int) (*chunk.SubChunk, int) {
	key := bedrock_level_operation_define.Index(dm, define.ChunkPos{position[0], position[2]})

	subChunkData, _, _ := db.DB.Get(append(key, bedrock_level_operation_define.KeySubChunkData, byte(int(position[1]))))
	if len(subChunkData) == 0 {
		return nil, 0
	}

	idx, err := QuickGetSubChunkIndex(subChunkData, dm.RangeUpperInclude())
	if err != nil {
		panic("? should not happened")
	}

	if idx != keyIdx {
		subChunk, idx, err := bedrock_level_operation_marshal.DecodeSubChunkBlocksNoNBT(subChunkData, dm.RangeUpperInclude())
		if err != nil {
			panic("?x2 should not happened")
		}
		return subChunk, idx
	}

	return nil, 0
}

func fixThisDay(wg *sync.WaitGroup) {
	defer wg.Done()

	mcdb, err := bedrock_level.New("thisDay", opt.DefaultCompression, false, nil)
	if err != nil {
		panic(err)
	}

	p := mcdb.(*bedrock_level_provider.Provider)

	p.IterAll(func(ck bedrock_level_operation.CanKV) (continueIter bool) {
		key := ck.Key()
		keyString := string(key)

		if strings.Contains(string(key), bedrock_level_operation_define.KeyBlobHash) {
			var dim uint32 = 0

			keyString = strings.ReplaceAll(keyString, bedrock_level_operation_define.KeyBlobHash, "")
			key = []byte(keyString)

			x := binary.LittleEndian.Uint32(key[0:4])
			z := binary.LittleEndian.Uint32(key[4:8])
			if len(key) > 8 {
				dim = binary.LittleEndian.Uint32(key[8:])
			}

			cp := define.ChunkPos([2]int32{int32(x), int32(z)})
			dm := define.Dimension(dim)

			if x > 1875000 || z > 1875000 {
				pterm.Error.Println("???a", x, z, dim)
				if p.Delete(ck.Key()) != nil {
					panic("aaaa?")
				}
				return true
			}

			for i := -4; i <= 23; i++ {
				c, idx := LoadSubChunk(p, dm, protocol.SubChunkPos{int32(x), int32(i), int32(z)}, i)
				if c == nil {
					continue
				}
				pterm.Success.Println("fix", x, z, dm)
				if mcdb.SaveSubChunk(dm, protocol.SubChunkPos{int32(x), int32(idx), int32(z)}, c) != nil {
					panic("??? should not happened")
				}
			}

			hashes := mcdb.LoadFullSubChunkBlobHash(dm, cp)
			for index := range hashes {
				posy := hashes[index].PosY
				hashes[index].PosY = int8((int32(posy)<<4 + int32(dm.RangeUpperInclude()[0])) >> 4)
			}
			if mcdb.SaveFullSubChunkBlobHash(dm, cp, hashes) != nil {
				panic("error save blob hash")
			}
			pterm.Warning.Println("update hash", x, z, dm, hashes)
		}

		return true
	})

	err = p.Close(false)
	if err != nil {
		panic(err)
	}
}

func fixDayAgo(wg *sync.WaitGroup) {
	defer wg.Done()

	mcdb, err := bedrock_level.New("dayAgo", opt.DefaultCompression, false, nil)
	if err != nil {
		panic(err)
	}

	p := mcdb.(*bedrock_level_provider.Provider)

	p.IterAll(func(ck bedrock_level_operation.CanKV) (continueIter bool) {
		key := ck.Key()
		keyString := string(key)

		if strings.Contains(string(key), bedrock_level_operation_define.KeyDeltaUpdate) {
			var dim uint32 = 0

			keyString = strings.ReplaceAll(keyString, bedrock_level_operation_define.KeyDeltaUpdate, "")
			key = []byte(keyString)

			x := binary.LittleEndian.Uint32(key[0:4])
			z := binary.LittleEndian.Uint32(key[4:8])
			if len(key) > 8 {
				dim = binary.LittleEndian.Uint32(key[8:])
			}

			cp := define.ChunkPos([2]int32{int32(x), int32(z)})
			dm := define.Dimension(dim)

			if x > 1875000 || z > 1875000 {
				pterm.Error.Println("???b", x, z, dim)
				if p.Delete(ck.Key()) != nil {
					panic("vvvvv?")
				}
				return true
			}

			if mcdb.SaveDeltaUpdate(dm, cp, nil) != nil {
				panic("a1")
			}
			if mcdb.SaveDeltaUpdateTimeStamp(dm, cp, 0) != nil {
				panic("a2")
			}

			indexKey := bedrock_level_operation_define.Index(dm, cp)
			if p.Delete(append(indexKey, bedrock_level_operation_define.KeyVersion)) != nil {
				panic("a3")
			}
			if p.Delete(append(indexKey, bedrock_level_operation_define.KeyFinalisation)) != nil {
				panic("a4")
			}

			for i := -4; i <= 23; i++ {
				subChunkKey := append(indexKey, bedrock_level_operation_define.KeySubChunkData, byte(i))
				if p.Delete(subChunkKey) != nil {
					panic("a5")
				}
			}

			pterm.Info.Println("clean delta update", x, z, dm)
		}

		return true
	})

	err = p.Close(false)
	if err != nil {
		panic(err)
	}
}

func main() {
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go fixThisDay(wg)
	go fixDayAgo(wg)
	wg.Wait()
	pterm.Success.Println("ALL FIXED :)")
}
