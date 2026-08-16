/**
 * SPDX-FileComment: Medical Devices Backend
 * SPDX-FileType: SOURCE
 * SPDX-FileContributor: ZHENG Robert
 * SPDX-FileCopyrightText: 2026 ZHENG Robert
 * SPDX-License-Identifier: Apache-2.0
 *
 * @file crypto.go
 * @brief AES-256-GCM encryption/decryption with Argon2id key derivation
 * @version 0.1.0
 * @date 2026-08-16
 *
 * @author ZHENG Robert (robert@hase-zheng.net)
 * @copyright Copyright (c) 2026 ZHENG Robert
 * @LICENSE Apache-2.0
 */

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	ArgonTime      = 3
	ArgonMemory    = 64 * 1024 // 64 MB
	ArgonThreads   = 1
	ArgonKeyLength = 32
)

func DeriveKey(password []byte, salt []byte) []byte {
	return argon2.IDKey(password, salt, ArgonTime, ArgonMemory, ArgonThreads, ArgonKeyLength)
}

func Encrypt(plaintext []byte, password []byte) ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	key := DeriveKey(password, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	result := make([]byte, len(salt)+len(nonce)+len(ciphertext))
	copy(result[0:16], salt)
	copy(result[16:16+len(nonce)], nonce)
	copy(result[16+len(nonce):], ciphertext)

	return result, nil
}

func Decrypt(data []byte, password []byte) ([]byte, error) {
	if len(data) < 16+12 {
		return nil, errors.New("invalid encrypted data size")
	}

	salt := data[:16]
	nonce := data[16 : 16+12]
	ciphertext := data[16+12:]

	key := DeriveKey(password, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, nonce, ciphertext, nil)
}

func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, ArgonTime, ArgonMemory, ArgonThreads, ArgonKeyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, ArgonMemory, ArgonTime, ArgonThreads, b64Salt, b64Hash), nil
}

func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}
	if parts[1] != "argon2id" {
		return false, errors.New("incompatible variant")
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, err
	}
	if version != argon2.Version {
		return false, errors.New("incompatible version")
	}

	var memory uint32
	var time uint32
	var threads uint8
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	compareHash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(hash)))

	if subtle.ConstantTimeCompare(hash, compareHash) == 1 {
		return true, nil
	}
	return false, nil
}
