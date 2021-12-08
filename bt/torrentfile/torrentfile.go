package torrentfile

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/jackpal/bencode-go"
	"ipaas_bwstress/bt/p2p"
	"ipaas_bwstress/bt/peers"
	"ipaas_bwstress/util"
	"ipaas_bwstress/util/collect"
	"ipaas_bwstress/util/config"
	. "ipaas_bwstress/util/log"
	"ipaas_bwstress/util/request"
	"net"
	"os"
	"time"
)

// Port to listen on
const Port uint16 = 6881

var DoneCh chan struct{}

// TorrentFile encodes the metadata from a .torrent file
type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

var TFiles = []string{
	"debian-11.0.0-i386-DVD-1.iso.torrent",
	"ubuntu-20.10-desktop-amd64.iso.torrent",
	"ubuntu-20.04.3-live-server-amd64.iso.torrent",
	"ubuntu-21.04-live-server-amd64.iso.torrent",
	"ubuntu-21.04-desktop-amd64.iso.torrent",
}
var TFilesPeers = make(map[string][]peers.Peer)

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		Ip   string `json:"ip"`
		Port uint16 `json:"port"`
	}
}

func getPeers(fname string) (peersList []peers.Peer) {
	respBody, httpCode, err := request.Get("https://ipaas.paigod.work/peer?file=" + fname)
	if err != nil {
		L.Errorf("Get %v peers: %v", fname, err.Error())
		return
	}

	if httpCode != 200 {
		L.Errorf("Get %v peers error, httpCode: %v", fname, httpCode)
		return
	}

	var resp Response
	if err = json.Unmarshal(respBody, &resp); err != nil {
		L.Errorf("Get %v peers error, json.Unmarshal %v", fname, err.Error())
		return
	}

	for _, p := range resp.Data {
		peersList = append(peersList, peers.Peer{IP: net.ParseIP(p.Ip), Port: p.Port})
	}

	L.Infof("Request %v peers success: %v peers", fname, len(resp.Data))
	return
}

func RequestTFilesPeers() {
	for {
		select {
		case <-DoneCh:
			return
		default:
			for _, fname := range TFiles {
				peerList := getPeers(fname)
				TFilesPeers[fname] = peerList
			}
			time.Sleep(5 * time.Minute)
		}
	}
}

// DownloadToFile downloads a torrent and writes it to a file
func (t *TorrentFile) DownloadToFile(ethName string, fname string) {
	torrent := p2p.Torrent{
		Peers:       make(chan peers.Peer),
		InfoHash:    t.InfoHash,
		PieceHashes: t.PieceHashes,
		PieceLength: t.PieceLength,
		Length:      t.Length,
		Name:        t.Name,
	}

	go func(ethName string) {
		for {
			select {
			case <-DoneCh:
				close(torrent.Peers)
				L.Infof("EthName %v p2p task exit", ethName)
				return
			default:
				time.Sleep(time.Minute)
				if config.LocalCfg.Enabled == 0 {
					continue
				}

				maxDownload := float64(config.LocalCfg.MaxDownload)
				maxRecvSendRate := float64(config.LocalCfg.RecvSendRate)

				//maxRecvSendBwPer 小于等于0 表示不判断总下行占总上行比例, 否则需要判断.
				if (maxRecvSendRate <= 0 && collect.OutEthRecvByteAvg < maxDownload) || (maxRecvSendRate > 0 && collect.OutEthRecvSendUseRate < maxRecvSendRate && collect.OutEthRecvByteAvg < maxDownload) {
					for ethN, outEthIfI := range collect.Net.IfiMap {
						outEthIfRecv := util.FormatFloat64(util.ByteToBitM(outEthIfI.RecvByteAvg))
						if ethN == ethName && outEthIfRecv < collect.OutIfMaxRecvBw {
							L.Infof("Add peers ethName %v, ip: %v, recv: %vMpbs, maxRecv: %vMpbs", ethName, outEthIfI.Ip, outEthIfRecv, collect.OutIfMaxRecvBw)
							for _, p := range TFilesPeers[fname] {
								torrent.Peers <- peers.Peer{IP: net.ParseIP(p.IP.String()), Port: p.Port}
							}
						}
					}
				}
			}
		}
	}(ethName)

	go torrent.Download(ethName, DoneCh)
}

// Open parses a torrent file
func Open(path string) (TorrentFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return TorrentFile{}, err
	}
	defer file.Close()

	bto := bencodeTorrent{}
	err = bencode.Unmarshal(file, &bto)
	if err != nil {
		return TorrentFile{}, err
	}
	return bto.toTorrentFile()
}

func (i *bencodeInfo) hash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, *i)
	if err != nil {
		return [20]byte{}, err
	}
	h := sha1.Sum(buf.Bytes())
	return h, nil
}

func (i *bencodeInfo) splitPieceHashes() ([][20]byte, error) {
	hashLen := 20 // Length of SHA-1 hash
	buf := []byte(i.Pieces)
	if len(buf)%hashLen != 0 {
		err := fmt.Errorf("Received malformed pieces of length %d", len(buf))
		return nil, err
	}
	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}
	return hashes, nil
}

func (bto *bencodeTorrent) toTorrentFile() (TorrentFile, error) {
	infoHash, err := bto.Info.hash()
	if err != nil {
		return TorrentFile{}, err
	}
	pieceHashes, err := bto.Info.splitPieceHashes()
	if err != nil {
		return TorrentFile{}, err
	}
	t := TorrentFile{
		Announce:    bto.Announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bto.Info.PieceLength,
		Length:      bto.Info.Length,
		Name:        bto.Info.Name,
	}
	return t, nil
}
