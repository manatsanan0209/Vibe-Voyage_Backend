package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

func (s *roomService) CreateInviteCode(ctx context.Context, roomID, creatorUserID uint, access string, expireTime time.Time) (*domain.RoomInviteCode, error) {
	if err := s.ensureRoomOwner(ctx, roomID, creatorUserID); err != nil {
		return nil, err
	}

	normalizedAccess := strings.ToLower(strings.TrimSpace(access))
	if normalizedAccess == "" {
		normalizedAccess = domain.InviteAccessView
	}
	if normalizedAccess != domain.InviteAccessView && normalizedAccess != domain.InviteAccessEdit {
		return nil, errors.New("access must be view or edit")
	}

	if expireTime.IsZero() {
		expireTime = time.Now().Add(24 * time.Hour)
	}
	if !expireTime.After(time.Now()) {
		return nil, errors.New("expire_time must be in the future")
	}

	code, err := s.generateUniqueInviteCode(ctx)
	if err != nil {
		return nil, err
	}

	invite := &domain.RoomInviteCode{
		RoomID:              roomID,
		InviteCodeCreatorID: creatorUserID,
		InviteCode:          code,
		Access:              normalizedAccess,
		ExpireTime:          expireTime,
	}
	if err := s.inviteRepo.Create(ctx, invite); err != nil {
		return nil, err
	}

	return invite, nil
}

func (s *roomService) JoinByInviteCode(ctx context.Context, userID uint, inviteCode string) (*domain.RoomMember, error) {
	code := strings.ToUpper(strings.TrimSpace(inviteCode))
	if code == "" {
		return nil, errors.New("invite_code is required")
	}

	invite, err := s.inviteRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if invite == nil {
		return nil, errors.New("invite code not found")
	}
	if invite.ExpireTime.Before(time.Now()) {
		return nil, errors.New("invite code has expired")
	}

	return s.AddMember(ctx, invite.RoomID, userID)
}

func (s *roomService) ListInviteCodes(ctx context.Context, roomID, requesterUserID uint) ([]domain.RoomInviteCode, error) {
	if err := s.ensureRoomOwner(ctx, roomID, requesterUserID); err != nil {
		return nil, err
	}

	invites, err := s.inviteRepo.ListByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	active := make([]domain.RoomInviteCode, 0, len(invites))
	for _, inv := range invites {
		if inv.ExpireTime.After(now) {
			active = append(active, inv)
		}
	}

	return active, nil
}

func (s *roomService) ensureRoomOwner(ctx context.Context, roomID, requesterUserID uint) error {
	members, err := s.memberRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		return err
	}

	for _, m := range members {
		if m.UserID == requesterUserID && m.Role == domain.RoleOwner {
			return nil
		}
	}

	return errors.New("only the room owner can perform this action")
}

func (s *roomService) generateUniqueInviteCode(ctx context.Context) (string, error) {
	const (
		alphabet         = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
		codeLength       = 8
		maxCodeGenTrials = 10
	)

	for i := 0; i < maxCodeGenTrials; i++ {
		code, err := randomCode(alphabet, codeLength)
		if err != nil {
			return "", err
		}

		exists, err := s.inviteRepo.ExistsCode(ctx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}

	return "", errors.New("failed to generate unique invite code")
}

func randomCode(alphabet string, length int) (string, error) {
	b := make([]byte, length)
	max := big.NewInt(int64(len(alphabet)))
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("failed to generate random code: %w", err)
		}
		b[i] = alphabet[n.Int64()]
	}
	return string(b), nil
}
