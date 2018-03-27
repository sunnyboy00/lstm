package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/owulveryck/lstm"
	"github.com/owulveryck/lstm/datasetter/char"
	G "gorgonia.org/gorgonia"
)

const (
	runes = `b«ME'àésèêüivOùquHÉa-A!ÇnJjçTByVepY,?xôXCmïW wfFU(gLNQ»R:°dPDîIktrcz.Shloâ)Gû
`
	filename = "../../data/tontons/input.txt"
)

var asRunes = []rune(runes)

func runeToIdx(r rune) (int, error) {
	for i := range asRunes {
		if asRunes[i] == r {
			return i, nil
		}
	}
	return 0, fmt.Errorf("Rune %v is not part of the vocabulary", string(r))
}

func idxToRune(i int) (rune, error) {
	var rn rune
	if i >= len([]rune(runes)) {
		return rn, fmt.Errorf("index invalid, no rune references")
	}
	return []rune(runes)[i], nil
}

func main() {
	vocabSize := len([]rune(runes))
	model := lstm.NewModel(vocabSize, vocabSize, vocabSize)
	learnrate := 0.001
	l2reg := 1e-6
	clipVal := float64(5)
	solver := G.NewRMSPropSolver(G.WithLearnRate(learnrate), G.WithL2Reg(l2reg), G.WithClip(clipVal))

	for i := 0; i < 100; i++ {
		f, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		tset := char.NewTrainingSet(f, runeToIdx, vocabSize, 35, 1)
		pause := make(chan struct{})
		infoChan, errc := model.Train(context.TODO(), tset, solver, pause)
		iter := 1
		for infos := range infoChan {
			if iter%100 == 0 {
				fmt.Printf("%v\n", infos)
			}
			if iter%500 == 0 {
				fmt.Println("\nGoing to predict")
				pause <- struct{}{}
				prediction := char.NewPrediction("Monsieur", runeToIdx, 50, vocabSize)
				model.Predict(context.TODO(), prediction)
				for _, node := range prediction.GetComputedVectors() {
					output := node.Value().Data().([]float32)
					max := float32(0)
					idx := 0
					for i := range output {
						if output[i] >= max {
							max = output[i]
							idx = i
						}
					}
					rne, err := idxToRune(idx)
					if err != nil {
						log.Fatal(err)
					}
					fmt.Printf(string(rne))
				}
				pause <- struct{}{}
			}
			iter++
		}
		err = <-errc
		if err == io.EOF {
			close(pause)
			return
		}
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		f.Close()
	}

}
