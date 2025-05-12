package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/beevik/etree"
)

const (
	cbrURL        = "https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx"
	soapAction    = "http://web.cbr.ru/KeyRate"
	dateLayout    = "2006-01-02"
	defaultMargin = 5.0
)

type CBRService struct{}

func NewCBRService() *CBRService {
	return &CBRService{}
}

func (s *CBRService) buildSOAPRequest() string {
	from := time.Now().AddDate(0, 0, -30).Format(dateLayout)
	to := time.Now().Format(dateLayout)
	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap12:Envelope xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
  <soap12:Body>
    <KeyRate xmlns="http://web.cbr.ru/">
      <fromDate>%s</fromDate>
      <ToDate>%s</ToDate>
    </KeyRate>
  </soap12:Body>
</soap12:Envelope>`, from, to)
}

func (s *CBRService) sendRequest(soapReq string) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", cbrURL, bytes.NewBufferString(soapReq))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	req.Header.Set("SOAPAction", soapAction)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ЦБ РФ: ошибка запроса: %w", err)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (s *CBRService) parseXML(raw []byte) (float64, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(raw); err != nil {
		return 0, fmt.Errorf("ЦБ РФ: ошибка парсинга XML: %w", err)
	}
	elems := doc.FindElements("//diffgram/KeyRate/KR")
	if len(elems) == 0 {
		return 0, errors.New("ЦБ РФ: данные по ставке не найдены")
	}
	rateEl := elems[0].FindElement("./Rate")
	if rateEl == nil {
		return 0, errors.New("ЦБ РФ: тег Rate отсутствует")
	}
	var rate float64
	if _, err := fmt.Sscanf(rateEl.Text(), "%f", &rate); err != nil {
		return 0, fmt.Errorf("ЦБ РФ: конвертация ставки: %w", err)
	}
	return rate, nil
}

func (s *CBRService) GetRate() (float64, error) {
	soap := s.buildSOAPRequest()
	raw, err := s.sendRequest(soap)
	if err != nil {
		return 0, err
	}
	rate, err := s.parseXML(raw)
	if err != nil {
		return 0, err
	}
	return rate + defaultMargin, nil
}
