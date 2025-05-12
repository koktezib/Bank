package service

import (
	"Bank/internal/model"
	"Bank/internal/repository"
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
)

var ErrCardNotYours = errors.New("account does not belong to user")

type CardService struct {
	pubKeyPath        string
	privKeyPath       string
	privKeyPassphrase string
	hmacSecret        []byte
	cardRepo          repository.CardRepository
	acctRepo          repository.AccountRepository
}

func NewCardService(
	pubKeyPath, privKeyPath, privKeyPassphrase, hmacSecret string,
	cr repository.CardRepository,
	ar repository.AccountRepository,
) *CardService {
	return &CardService{
		pubKeyPath:        pubKeyPath,
		privKeyPath:       privKeyPath,
		privKeyPassphrase: privKeyPassphrase,
		hmacSecret:        []byte(hmacSecret),
		cardRepo:          cr,
		acctRepo:          ar,
	}
}

func (s *CardService) GenerateCard(userID, accountID int) (*model.Card, error) {
	acc, err := s.acctRepo.GetByID(accountID)
	if err != nil {
		return nil, err
	}
	if acc.UserID != userID {
		return nil, ErrCardNotYours
	}

	number := generateLuhnNumber(16)

	expiry := time.Now().AddDate(3, 0, 0).Format("01/2006") // MM/YYYY

	cvv := fmt.Sprintf("%03d", randInt(0, 999))

	numEnc, err := encryptWithPGP(s.pubKeyPath, []byte(number))
	if err != nil {
		return nil, err
	}
	expEnc, err := encryptWithPGP(s.pubKeyPath, []byte(expiry))
	if err != nil {
		return nil, err
	}

	cvvHash, err := bcrypt.GenerateFromPassword([]byte(cvv), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	h := hmac.New(sha256.New, s.hmacSecret)
	h.Write([]byte(number))
	mac := hex.EncodeToString(h.Sum(nil))

	card := &model.Card{
		AccountID:       accountID,
		NumberEncrypted: numEnc,
		ExpiryEncrypted: expEnc,
		CVVHash:         string(cvvHash),
		HMAC:            mac,
	}

	if err := s.cardRepo.Create(card); err != nil {
		return nil, err
	}
	return card, nil
}

func (s *CardService) ListCards(userID int) ([]*model.CardResponse, error) {
	accounts, err := s.acctRepo.ListByUser(userID)
	if err != nil {
		return nil, err
	}

	var out []*model.CardResponse
	for _, acc := range accounts {
		cards, err := s.cardRepo.ListByAccount(acc.ID)
		if err != nil {
			return nil, err
		}
		for _, c := range cards {
			numPlain, err := decryptWithPGP(s.privKeyPath, s.privKeyPassphrase, c.NumberEncrypted)
			if err != nil {
				return nil, err
			}
			expPlain, err := decryptWithPGP(s.privKeyPath, s.privKeyPassphrase, c.ExpiryEncrypted)
			if err != nil {
				return nil, err
			}
			out = append(out, &model.CardResponse{
				ID:        c.ID,
				Number:    string(numPlain),
				Expiry:    string(expPlain),
				CreatedAt: c.CreatedAt,
			})
		}
	}
	return out, nil
}

func randInt(min, max int) int {
	n, _ := rand.Int(rand.Reader,
		big.NewInt(int64(max-min+1)))
	return int(n.Int64()) + min
}

func generateLuhnNumber(length int) string {
	digits := make([]int, length)
	for i := 0; i < length-1; i++ {
		digits[i] = randInt(0, 9)
	}
	sum := 0
	for i := 0; i < length-1; i++ {
		d := digits[length-2-i]
		if i%2 == 0 {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	check := (10 - (sum % 10)) % 10
	digits[length-1] = check
	s := ""
	for _, d := range digits {
		s += strconv.Itoa(d)
	}
	return s
}

func encryptWithPGP(pubKeyPath string, data []byte) ([]byte, error) {
	f, err := os.Open(pubKeyPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	entityList, err := openpgp.ReadArmoredKeyRing(f)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	w, err := armor.Encode(buf, "PGP MESSAGE", nil)
	if err != nil {
		return nil, err
	}
	pt, err := openpgp.Encrypt(w, entityList, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	if _, err := pt.Write(data); err != nil {
		return nil, err
	}
	pt.Close()
	w.Close()
	return buf.Bytes(), nil
}

func decryptWithPGP(privKeyPath, passphrase string, cipher []byte) ([]byte, error) {
	keyFile, err := os.Open(privKeyPath)
	if err != nil {
		return nil, fmt.Errorf("open privKey: %w", err)
	}
	defer keyFile.Close()

	entityList, err := openpgp.ReadArmoredKeyRing(keyFile)
	if err != nil {
		return nil, fmt.Errorf("read key ring: %w", err)
	}
	if len(entityList) == 0 {
		return nil, errors.New("no PGP entities found")
	}

	for _, ent := range entityList {
		if ent.PrivateKey != nil && ent.PrivateKey.Encrypted {
			if err := ent.PrivateKey.Decrypt([]byte(passphrase)); err != nil {
				return nil, fmt.Errorf("decrypt privKey: %w", err)
			}
		}
		for _, sub := range ent.Subkeys {
			if sub.PrivateKey != nil && sub.PrivateKey.Encrypted {
				if err := sub.PrivateKey.Decrypt([]byte(passphrase)); err != nil {
					return nil, fmt.Errorf("decrypt subkey: %w", err)
				}
			}
		}
	}

	block, err := armor.Decode(bytes.NewReader(cipher))
	if err != nil {
		return nil, fmt.Errorf("armor decode: %w", err)
	}

	promptFunc := func(keys []openpgp.Key, symmetric bool) ([]byte, error) {
		return []byte(passphrase), nil
	}
	md, err := openpgp.ReadMessage(block.Body, entityList, promptFunc, nil)
	if err != nil {
		return nil, fmt.Errorf("read message: %w", err)
	}

	plain, err := io.ReadAll(md.UnverifiedBody)
	if err != nil {
		return nil, fmt.Errorf("read plaintext: %w", err)
	}
	return plain, nil
}
