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

// DownloadToFile downloads a torrent and writes it to a file
func (t *TorrentFile) DownloadToFile(ethName string, tFile string) error {
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
				maxDownload := float64(config.LocalCfg.MaxDownload)
				maxRecvSendRate := float64(config.LocalCfg.RecvSendRate)

				//maxRecvSendBwPer 小于等于0 表示不判断总下行占总上行比例, 否则需要判断.
				if (maxRecvSendRate <= 0 && collect.OutEthRecvByteAvg < maxDownload) || (maxRecvSendRate > 0 && collect.OutEthRecvSendUseRate < maxRecvSendRate && collect.OutEthRecvByteAvg < maxDownload) {
					for ethN, outEthIfI := range collect.Net.IfiMap {
						outEthIfRecv := util.FormatFloat64(util.ByteToBitM(outEthIfI.RecvByteAvg))
						if ethN == ethName && outEthIfRecv < collect.OutIfMaxRecvBw {
							L.Infof("Add peers ethName %v, ip: %v, recv: %vMpbs, maxRecv: %vMpbs", ethName, outEthIfI.Ip, outEthIfRecv, collect.OutIfMaxRecvBw)
							respBody, httpCode, err := request.Get("https://ipaas.paigod.work/peer?file=" + tFile)
							if err != nil {
								L.Errorf("EthName: %v, ip: %v, requestPeers: %v", ethName, outEthIfI.Ip, err.Error())
								break
							}

							if httpCode != 200 {
								L.Errorf("EthName: %v, ip: %v, requestPeers: %v", ethName, outEthIfI.Ip, err.Error())
								break
							}

							type Resp struct {
								Code int    `json:"code"`
								Msg  string `json:"msg"`
								Data []struct {
									Ip   string `json:"ip"`
									Port uint16 `json:"port"`
								}
							}
							var resp Resp
							if err = json.Unmarshal(respBody, &resp); err != nil {
								L.Errorf("json.Unmarshal %v", err.Error())
								break
							}

							L.Infof("Request peers success: %v peers", len(resp.Data))
							for _, peer := range resp.Data {
								p := peers.Peer{IP: net.ParseIP(peer.Ip), Port: peer.Port}
								torrent.Peers <- p
							}
						}
					}
				}
			}
		}
	}(ethName)

	err := torrent.Download(ethName, DoneCh)
	if err != nil {
		return err
	}
	return nil
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
