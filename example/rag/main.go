package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/showntop/llmack/rag"
	"github.com/showntop/llmack/vdb"
	vredis "github.com/showntop/llmack/vdb/redis"
)

func main() {
	ctx := context.Background()

	config := &vredis.Config{}
	config.Addr = "127.0.0.1:6379"
	config.Password = "cdgxxx2025@tx"
	config.DB = 0
	config.Index = "vdb"
	// config = 1536
	config.FieldSchema = []*redis.FieldSchema{
		{},
	}

	indexer, err := rag.NewIndexer(vredis.Name, config)
	if err != nil {
		panic(err)
	}
	indexer.Index(ctx, mockDocuments(), &rag.Options{
		LibraryID:      1,
		Kind:           "text",
		IndexID:        1,
		TopK:           10,
		ScoreThreshold: 0.5,
	})

	entities, err := indexer.Retrieve(ctx, "你好", &rag.Options{
		LibraryID:      1,
		Kind:           "text",
		IndexID:        1,
		TopK:           10,
		ScoreThreshold: 0.5,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(entities)
}

func mockDocuments() []*vdb.Document {
	contents := `
1. 埃菲尔铁塔：位于法国巴黎，是世界上最著名的地标之一，由古斯塔夫·埃菲尔设计，建于1889年。
2. 长城：位于中国，是世界七大奇迹之一，始建于秦朝至明朝，全长超过20000公里。
3. 大峡谷国家公园：位于美国亚利桑那州，以其深深的峡谷和壮丽的景色而闻名，它被科罗拉多河切割而成。
4. 罗马斗兽场：位于意大利罗马，建于公元70-80年间，是古罗马帝国最大的圆形竞技场。
5. 泰姬陵：位于印度阿格拉，由莫卧儿皇帝沙贾汗于1653年为纪念他的妻子而建成，是世界新七大奇迹之一。
6. 悉尼歌剧院：位于澳大利亚悉尼港，是20世纪最具标志性的建筑之一，以其独特的帆船设计而闻名。
7. 卢浮宫博物馆：位于法国巴黎，是世界上最大的博物馆之一，收藏丰富，包括达·芬奇的《蒙娜丽莎》和希腊的《米洛的维纳斯》。
8. 尼亚加拉大瀑布：位于美国和加拿大边境，由三个主要瀑布组成，其壮观的景色每年吸引数百万游客。
9. 圣索菲亚大教堂：位于土耳其伊斯坦布尔，最初建于公元537年，曾是东正教大教堂和清真寺，现为博物馆。
10. 马丘比丘：位于秘鲁安第斯山脉高原上的古印加遗址，世界新七大奇迹之一，海拔2400多米。
 `

	var docs []*vdb.Document
	for idx, str := range strings.Split(contents, "\n") {
		if str == "" {
			continue
		}
		docs = append(docs, &vdb.Document{
			ID:      strconv.FormatInt(int64(idx+1), 10),
			Content: str,
		})
	}
	return docs
}
