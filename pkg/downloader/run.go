package downloader

func (d *downloader) Run(plutoChTx chan<- string, plutoChRx <-chan string) {
	masterPlaylistStopCh := make(chan struct{})
	workerChTx := make(chan string, 2)
	workerChRx := make(chan string, 2)

	d.masterPlaylistStopCh = masterPlaylistStopCh

	go d.masterPlaylistWorker(plutoChTx, workerChTx, plutoChRx, workerChRx)

	// precisa inverter o Tx e Rx para que
	// o Rx de um seja o Tx do outro
	go d.mediaPlaylistWorker(workerChRx, workerChTx)
}
