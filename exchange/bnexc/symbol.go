package bnexc

import (
	"github.com/go-gotop/kit/exchange"
)

var (
 bnSymbols = map[string]string{
  exchange.BTCUSDT: "BTCUSDT",
  exchange.BNBUSDT: "BNBUSDT",
  exchange.DOGEUSDT: "DOGEUSDT",
  exchange.FETUSDT: "FETUSDT",
  exchange.LINKUSDT: "LINKUSDT",
  exchange.ORDIUSDT: "ORDIUSDT",
  exchange.CKBUSDT: "CKBUSDT",
  exchange.SOLUSDT: "SOLUSDT",
  exchange.ETHUSDT: "ETHUSDT",
  exchange.AGIXUSDT: "AGIXUSDT",
  exchange.XRPUSDT: "XRPUSDT",
 }
 bnAsstes = map[string]string{
  exchange.BTC: "BTC",
  exchange.BNB: "BNB",
  exchange.DOGE: "DOGE",
  exchange.FET: "FET",
  exchange.LINK: "LINK",
  exchange.ORDI: "ORDI",
  exchange.USDT: "USDT",
  exchange.CKB: "CKB",
  exchange.SOL: "SOL",
  exchange.ETH: "ETH",
  exchange.AGIX: "AGIX",
  exchange.XRP: "XRP",
 }

 reverseBnSymbols = map[string]string{
  "BTCUSDT": exchange.BTCUSDT,
  "BNBUSDT": exchange.BNBUSDT,
  "DOGEUSDT": exchange.DOGEUSDT,
  "FETUSDT": exchange.FETUSDT,
  "LINKUSDT": exchange.LINKUSDT,
  "ORDIUSDT": exchange.ORDIUSDT,
  "CKBUSDT": exchange.CKBUSDT,
  "SOLUSDT": exchange.SOLUSDT,
  "ETHUSDT": exchange.ETHUSDT,
  "AGIXUSDT": exchange.AGIXUSDT,
  "XRPUSDT": exchange.XRPUSDT,
 }
 reverseBnAssets = map[string]string{
  "BTC": exchange.BTC,
  "BNB": exchange.BNB,
  "DOGE": exchange.DOGE,
  "FET": exchange.FET,
  "LINK": exchange.LINK,
  "ORDI": exchange.ORDI,
  "USDT": exchange.USDT,
  "CKB": exchange.CKB,
  "SOL": exchange.SOL,
  "ETH": exchange.ETH,
  "AGIX": exchange.AGIX,
  "XRP": exchange.XRP,
 }
)
