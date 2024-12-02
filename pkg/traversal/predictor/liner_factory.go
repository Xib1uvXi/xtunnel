package predictor

type IncLinearPortPredictor struct {
	port  int
	count int
}

func NewIncLinearPortPredictor(port int) *IncLinearPortPredictor {
	return &IncLinearPortPredictor{port: port}
}

func (d *IncLinearPortPredictor) NextPort() int {
	if d.port >= 65535 {
		return 0
	}

	if d.count >= 230 {
		return 0
	}

	d.port++
	d.count++

	return d.port
}

type SubLinearPortPredictor struct {
	port  int
	count int
}

func NewSubLinearPortPredictor(port int) *SubLinearPortPredictor {
	return &SubLinearPortPredictor{port: port}
}

func (d *SubLinearPortPredictor) NextPort() int {
	if d.port <= 1024 {
		return 0
	}

	if d.count >= 230 {
		return 0
	}

	d.port--
	d.count++

	return d.port
}
