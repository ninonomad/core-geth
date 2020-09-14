package core

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

var yuckyGlobalTestEnableMess = false

func runMESSTest(t *testing.T, easyL, hardL, caN int, easyT, hardT int64) (hardHead bool, err error) {
	// Generate the original common chain segment and the two competing forks
	engine := ethash.NewFaker()

	db := rawdb.NewMemoryDatabase()
	genesis := params.DefaultMessNetGenesisBlock()
	genesisB := MustCommitGenesis(db, genesis)

	chain, err := NewBlockChain(db, nil, genesis.Config, engine, vm.Config{}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer chain.Stop()
	chain.EnableArtificialFinality(yuckyGlobalTestEnableMess)

	easy, _ := GenerateChain(genesis.Config, genesisB, engine, db, easyL, func(i int, b *BlockGen) {
		b.SetNonce(types.EncodeNonce(uint64(rand.Int63n(math.MaxInt64))))
		b.OffsetTime(easyT)
	})
	commonAncestor := easy[caN-1]
	hard, _ := GenerateChain(genesis.Config, commonAncestor, engine, db, hardL, func(i int, b *BlockGen) {
		b.SetNonce(types.EncodeNonce(uint64(rand.Int63n(math.MaxInt64))))
		b.OffsetTime(hardT)
	})

	if _, err := chain.InsertChain(easy); err != nil {
		t.Fatal(err)
	}
	_, err = chain.InsertChain(hard)
	hardHead = chain.CurrentBlock().Hash() == hard[len(hard)-1].Hash()
	return
}

func TestBlockChain_AF_ECBP1100(t *testing.T) {
	yuckyGlobalTestEnableMess = true
	defer func() {
		yuckyGlobalTestEnableMess = false
	}()

	cases := []struct {
		easyLen, hardLen, commonAncestorN int
		easyOffset, hardOffset            int64
		hardGetsHead, accepted            bool
	}{
		// Hard has insufficient total difficulty / length and is rejected.
		{
			5000, 7500, 2500,
			50, -9,
			false, false,
		},
		// Hard has sufficient total difficulty / length and is accepted.
		{
			1000, 7, 995,
			60, 0,
			true, true,
		},
		// Hard has sufficient total difficulty / length and is accepted.
		{
			1000, 7, 995,
			60, 7,
			true, true,
		},
		// Hard has sufficient total difficulty / length and is accepted.
		{
			1000, 1, 999,
			30, 1,
			true, true,
		},
		// Hard has sufficient total difficulty / length and is accepted.
		{
			500, 3, 497,
			0, -8,
			true, true,
		},
		// Hard has sufficient total difficulty / length and is accepted.
		{
			500, 4, 496,
			0, -9,
			true, true,
		},
		// Hard has sufficient total difficulty / length and is accepted.
		{
			500, 5, 495,
			0, -9,
			true, true,
		},
		// Hard has sufficient total difficulty / length and is accepted.
		{
			500, 6, 494,
			0, -9,
			true, true,
		},
		// Hard has sufficient total difficulty / length and is accepted.
		{
			500, 7, 493,
			0, -9,
			true, true,
		},
		// Hard has sufficient total difficulty / length and is accepted.
		{
			500, 8, 492,
			0, -9,
			true, true,
		},
		// Hard has sufficient total difficulty / length and is accepted.
		{
			500, 9, 491,
			0, -9,
			true, true,
		},
		// Hard has sufficient total difficulty / length and is accepted.
		{
			500, 12, 488,
			0, -9,
			true, true,
		},
		// Hard has sufficient total difficulty / length and is accepted.
		{
			500, 20, 480,
			0, -9,
			true, true,
		},
		// Hard has sufficient total difficulty / length and is accepted.
		{
			500, 40, 460,
			0, -9,
			true, true,
		},
		// Hard has sufficient total difficulty / length and is accepted.
		{
			500, 60, 440,
			0, -9,
			true, true,
		},
		// Hard has insufficient total difficulty / length and is rejected.
		{
			500, 200, 300,
			0, -9,
			false, false,
		},
		// Hard has insufficient total difficulty / length and is rejected.
		{
			500, 200, 300,
			7, -9,
			false, false,
		},
		// Hard has insufficient total difficulty / length and is rejected.
		{
			500, 200, 300,
			17, -9,
			false, false,
		},
		// Hard has sufficient total difficulty / length and is accepted.
		{
			500, 200, 300,
			47, -9,
			true, true,
		},
		// Hard has insufficient total difficulty / length and is rejected.
		{
			500, 200, 300,
			47, -8,
			false, false,
		},
		// Hard has insufficient total difficulty / length and is rejected.
		{
			500, 200, 300,
			17, -8,
			false, false,
		},
		// Hard has insufficient total difficulty / length and is rejected.
		{
			500, 200, 300,
			7, -8,
			false, false,
		},
		// Hard has insufficient total difficulty / length and is rejected.
		{
			500, 200, 300,
			0, -8,
			false, false,
		},
		// Hard has insufficient total difficulty / length and is rejected.
		{
			500, 100, 400,
			0, -7,
			false, false,
		},
		// Hard is accepted, but does not have greater total difficulty,
		// and is not set as the chain head.
		{
			1000, 1, 900,
			60, -9,
			false, true,
		},
		// Hard is shorter, but sufficiently heavier chain, is accepted.
		{
			500, 100, 390,
			60, -9,
			true, true,
		},
	}

	for i, c := range cases {
		hardHead, err := runMESSTest(t, c.easyLen, c.hardLen, c.commonAncestorN, c.easyOffset, c.hardOffset)
		if (err != nil && c.accepted) || (err == nil && !c.accepted) || (hardHead != c.hardGetsHead) {
			t.Errorf("case=%d [easy=%d hard=%d ca=%d eo=%d ho=%d] want.accepted=%v want.hardHead=%v got.hardHead=%v err=%v",
				i,
				c.easyLen, c.hardLen, c.commonAncestorN, c.easyOffset, c.hardOffset,
				c.accepted, c.hardGetsHead, hardHead, err)
		}
	}
}

func TestBlockChain_GenerateMESSPlot(t *testing.T) {
	t.Skip("This test plots graph of chain acceptance for visualization.")

	easyLen := 200
	maxHardLen := 100

	generatePlot := func(title, fileName string) {
		p, err := plot.New()
		if err != nil {
			log.Panic(err)
		}
		p.Title.Text = title
		p.X.Label.Text = "Block Depth"
		p.Y.Label.Text = "Relative Block Time Delta (10 seconds + y)"

		accepteds := plotter.XYs{}
		rejecteds := plotter.XYs{}
		sides := plotter.XYs{}

		for i := 1; i <= maxHardLen; i++ {
			for j := -9; j <= 8; j++ {
				fmt.Println("running", i, j)
				hardHead, err := runMESSTest(t, easyLen, i, easyLen-i, 0, int64(j))
				point := plotter.XY{X: float64(i), Y: float64(j)}
				if err == nil && hardHead {
					accepteds = append(accepteds, point)
				} else if err == nil && !hardHead {
					sides = append(sides, point)
				} else if err != nil {
					rejecteds = append(rejecteds, point)
				}

				if err != nil {
					t.Log(err)
				}
			}
		}

		scatterAccept, _ := plotter.NewScatter(accepteds)
		scatterReject, _ := plotter.NewScatter(rejecteds)
		scatterSide, _ := plotter.NewScatter(sides)

		pixelWidth := vg.Length(1000)

		scatterAccept.Color = color.RGBA{R: 152, G: 236, B: 161, A: 255}
		scatterAccept.Shape = draw.BoxGlyph{}
		scatterAccept.Radius = vg.Length((float64(pixelWidth) / float64(maxHardLen)) * 2 / 3)
		scatterReject.Color = color.RGBA{R: 236, G: 106, B: 94, A: 255}
		scatterReject.Shape = draw.BoxGlyph{}
		scatterReject.Radius = vg.Length((float64(pixelWidth) / float64(maxHardLen)) * 2 / 3)
		scatterSide.Color = color.RGBA{R: 190, G: 197, B: 236, A: 255}
		scatterSide.Shape = draw.BoxGlyph{}
		scatterSide.Radius = vg.Length((float64(pixelWidth) / float64(maxHardLen)) * 2 / 3)

		p.Add(scatterAccept)
		p.Legend.Add("Accepted", scatterAccept)
		p.Add(scatterReject)
		p.Legend.Add("Rejected", scatterReject)
		p.Add(scatterSide)
		p.Legend.Add("Sidechained", scatterSide)

		p.Legend.YOffs = -30

		err = p.Save(pixelWidth, 300, fileName)
		if err != nil {
			log.Panic(err)
		}
	}
	yuckyGlobalTestEnableMess = true
	defer func() {
		yuckyGlobalTestEnableMess = false
	}()
	baseTitle := fmt.Sprintf("Accept/Reject Reorgs: Relative Time (Difficulty) over Proposed Segment Length (%d-block original chain)", easyLen)
	generatePlot(baseTitle, "reorgs-MESS.png")
	yuckyGlobalTestEnableMess = false
	generatePlot("WITHOUT MESS: "+baseTitle, "reorgs-noMESS.png")
}

func TestEcbp1100AGSinusoidalA(t *testing.T) {
	cases := []struct {
		in, out float64
	}{
		{0, 1},
		{25132, 31},
	}
	tolerance := 0.0000001
	for i, c := range cases {
		if got := ecbp1100AGSinusoidalA(c.in); got < c.out-tolerance || got > c.out+tolerance {
			t.Fatalf("%d: in: %0.6f want: %0.6f got: %0.6f", i, c.in, c.out, got)
		}
	}
}

func TestBlockChain_AF_Difficulty(t *testing.T) {
	// Generate the original common chain segment and the two competing forks
	engine := ethash.NewFaker()

	db := rawdb.NewMemoryDatabase()
	genesis := params.DefaultMessNetGenesisBlock()
	genesis.Timestamp = 1
	genesisB := MustCommitGenesis(db, genesis)

	chain, err := NewBlockChain(db, nil, genesis.Config, engine, vm.Config{}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer chain.Stop()
	// chain.EnableArtificialFinality(yuckyGlobalTestEnableMess)

	cases := []struct {
		easyLen, hardLen, commonAncestorN int
		easyOffset, hardOffset            int64
		hardGetsHead, accepted            bool
	}{
		// {
		// 	1000, 800, 200,
		// 	10, 1,
		// 	true, true,
		// },
		// {
		// 	1000, 800, 200,
		// 	60, 1,
		// 	true, true,
		// },
		// {
		// 	10000, 8000, 2000,
		// 	60, 1,
		// 	true, true,
		// },
		// {
		// 	20000, 18000, 2000,
		// 	10, 1,
		// 	true, true,
		// },
		// {
		// 	20000, 18000, 2000,
		// 	60, 1,
		// 	true, true,
		// },
		// {
		// 	10000, 8000, 2000,
		// 	10, 20,
		// 	true, true,
		// },


		// {
		// 	1000, 1, 999,
		// 	10, 1,
		// 	true, true,
		// },
		// {
		// 	1000, 10, 990,
		// 	10, 1,
		// 	true, true,
		// },
		// {
		// 	1000, 100, 900,
		// 	10, 1,
		// 	true, true,
		// },
		// {
		// 	1000, 200, 800,
		// 	10, 1,
		// 	true, true,
		// },
		// {
		// 	1000, 500, 500,
		// 	10, 1,
		// 	true, true,
		// },
		// {
		// 	1000, 999, 1,
		// 	10, 1,
		// 	true, true,
		// },
		// {
		// 	5000, 4000, 1000,
		// 	10, 1,
		// 	true, true,
		// },

		{
			10000, 9000, 1000,
			10, 1,
			true, true,
		},

		{
			7000, 6500, 500,
			10, 1,
			true, true,
		},

		// {
		// 	100, 90, 10,
		// 	10, 1,
		// 	true, true,
		// },
	}

	// poissonTime := func(b *BlockGen, seconds int64) {
	// 	poisson := distuv.Poisson{Lambda: float64(seconds)}
	// 	r := poisson.Rand()
	// 	if r < 1 {
	// 		r = 1
	// 	}
	// 	if r > float64(seconds) * 1.5 {
	// 		r = float64(seconds)
	// 	}
	// 	chainreader := &fakeChainReader{config: b.config}
	// 	b.header.Time = b.parent.Time() + uint64(r)
	// 	b.header.Difficulty = b.engine.CalcDifficulty(chainreader, b.header.Time, b.parent.Header())
	// 	for err := b.engine.VerifyHeader(chainreader, b.header, false);
	// 		err != nil && err != consensus.ErrUnknownAncestor && b.header.Time > b.parent.Header().Time; {
	// 		t.Log(err)
	// 		r -= 1
	// 		b.header.Time = b.parent.Time() + uint64(r)
	// 		b.header.Difficulty = b.engine.CalcDifficulty(chainreader, b.header.Time, b.parent.Header())
	// 	}
	// }

	for i, c := range cases {

		chain.Reset()
		easy, _ := GenerateChain(genesis.Config, genesisB, engine, db, c.easyLen, func(i int, b *BlockGen) {
			b.SetNonce(types.EncodeNonce(uint64(rand.Int63n(math.MaxInt64))))
			// poissonTime(b, c.easyOffset)
			b.OffsetTime(c.easyOffset - 10)
		})
		commonAncestor := easy[c.commonAncestorN-1]
		hard, _ := GenerateChain(genesis.Config, commonAncestor, engine, db, c.hardLen, func(i int, b *BlockGen) {
			b.SetNonce(types.EncodeNonce(uint64(rand.Int63n(math.MaxInt64))))
			// poissonTime(b, c.hardOffset)
			b.OffsetTime(c.hardOffset - 10)
		})
		if _, err := chain.InsertChain(easy); err != nil {
			t.Fatal(err)
		}
		_, err := chain.InsertChain(hard)
		hardHead := chain.CurrentBlock().Hash() == hard[len(hard)-1].Hash()

		commons := plotter.XYs{}
		easys := plotter.XYs{}
		hards := plotter.XYs{}
		tdrs := plotter.XYs{}
		antigravities := plotter.XYs{}
		antigravities2 := plotter.XYs{}
		for i := 0; i < c.easyLen; i++ {
			td := chain.GetTd(easy[i].Hash(), easy[i].NumberU64())
			point := plotter.XY{X: float64(easy[i].NumberU64()), Y: float64(td.Uint64())}
			if i <= c.commonAncestorN {
				commons = append(commons, point)
			} else {
				easys = append(easys, point)
			}
		}
		// td ratios
		for i := 0; i < c.hardLen; i++ {
			td := chain.GetTd(hard[i].Hash(), hard[i].NumberU64())
			point := plotter.XY{X: float64(hard[i].NumberU64()), Y: float64(td.Uint64())}
			hards = append(hards, point)

			ee := c.commonAncestorN+i
			y := chain.getTDRatio(commonAncestor.Header(), easy[ee].Header(), hard[i].Header())

			ecbp := ecbp1100AGSinusoidalA(float64(hard[i].Header().Time - commonAncestor.Header().Time))
			ecbp2 := ecbp1100AGExpA(float64(hard[i].Header().Time - commonAncestor.Header().Time))
			t.Log(y, ecbp, ecbp2)

			tdrs = append(tdrs, plotter.XY{X: float64(hard[i].NumberU64()), Y: y})
			antigravities = append(antigravities, plotter.XY{X: float64(hard[i].NumberU64()), Y: ecbp})
			antigravities2 = append(antigravities2, plotter.XY{X: float64(hard[i].NumberU64()), Y: ecbp2})
		}
		scatterCommons, _ := plotter.NewScatter(commons)
		scatterEasys, _ := plotter.NewScatter(easys)
		scatterHards, _ := plotter.NewScatter(hards)

		scatterTDRs, _ := plotter.NewScatter(tdrs)
		scatterAntigravities, _ := plotter.NewScatter(antigravities)
		scatterAntigravities2, _ := plotter.NewScatter(antigravities2)

		scatterCommons.Color = color.RGBA{R: 190, G: 197, B: 236, A: 255}
		scatterCommons.Shape = draw.CircleGlyph{}
		scatterCommons.Radius = 2
		scatterEasys.Color = color.RGBA{R: 152, G: 236, B: 161, A: 255} // green
		scatterEasys.Shape = draw.CircleGlyph{}
		scatterEasys.Radius = 2
		scatterHards.Color = color.RGBA{R: 236, G: 106, B: 94, A: 255}
		scatterHards.Shape = draw.CircleGlyph{}
		scatterHards.Radius = 2

		p, err := plot.New()
		if err != nil {
			log.Panic(err)
		}
		p.Add(scatterCommons)
		p.Legend.Add("Commons", scatterCommons)
		p.Add(scatterEasys)
		p.Legend.Add("Easys", scatterEasys)
		p.Add(scatterHards)
		p.Legend.Add("Hards", scatterHards)
		p.Title.Text = fmt.Sprintf("TD easy=%d hard=%d", c.easyOffset, c.hardOffset)
		p.Save(1000, 600, fmt.Sprintf("plot-td-%d-%d-%d-%d-%d.png", c.easyLen, c.commonAncestorN, c.hardLen, c.easyOffset, c.hardOffset))

		p, err = plot.New()
		if err != nil {
			log.Panic(err)
		}
		scatterTDRs.Color = color.RGBA{R: 236, G: 106, B: 94, A: 255} // red
		scatterTDRs.Radius = 1
		scatterTDRs.Shape = draw.PyramidGlyph{}
		p.Add(scatterTDRs)

		scatterAntigravities.Color = color.RGBA{R: 190, G: 197, B: 236, A: 255} // blue
		scatterAntigravities.Radius = 1
		scatterAntigravities.Shape = draw.PlusGlyph{}
		p.Add(scatterAntigravities)

		scatterAntigravities2.Color = color.RGBA{R: 152, G: 236, B: 161, A: 255} // green
		scatterAntigravities2.Radius = 1
		scatterAntigravities2.Shape = draw.PlusGlyph{}
		p.Add(scatterAntigravities2)

		p.Title.Text = fmt.Sprintf("TD Ratio easy=%d hard=%d", c.easyOffset, c.hardOffset)
		p.Save(1000, 600, fmt.Sprintf("plot-td-ratio-%d-%d-%d-%d-%d.png", c.easyLen, c.commonAncestorN, c.hardLen, c.easyOffset, c.hardOffset))


		if (err != nil && c.accepted) || (err == nil && !c.accepted) || (hardHead != c.hardGetsHead) {
			t.Errorf("case=%d [easy=%d hard=%d ca=%d eo=%d ho=%d] want.accepted=%v want.hardHead=%v got.hardHead=%v err=%v",
				i,
				c.easyLen, c.hardLen, c.commonAncestorN, c.easyOffset, c.hardOffset,
				c.accepted, c.hardGetsHead, hardHead, err)
		}
	}

}
