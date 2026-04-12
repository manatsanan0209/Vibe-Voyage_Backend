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

type roomMemberService struct {
	memberRepo domain.RoomMemberRepository
	inviteRepo domain.RoomInviteCodeRepository
}

func NewRoomMemberService(memberRepo domain.RoomMemberRepository, inviteRepo domain.RoomInviteCodeRepository) domain.RoomMemberService {
	return &roomMemberService{memberRepo: memberRepo, inviteRepo: inviteRepo}
}

func (s *roomMemberService) GetMembersByRoomID(ctx context.Context, roomID uint) ([]domain.RoomMember, error) {
	return s.memberRepo.GetByRoomID(ctx, roomID)
}

func (s *roomMemberService) GetRoomsByUserID(ctx context.Context, userID uint) ([]domain.UserRoomSummary, error) {
	return s.memberRepo.GetRoomsByUserID(ctx, userID)
}

func (s *roomMemberService) DeleteMember(ctx context.Context, roomID, requesterUserID, roomMemberID uint) error {
	// ตรวจสอบว่า requester เป็น owner ของ room นี้
	members, err := s.memberRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		return err
	}

	isOwner := false
	for _, m := range members {
		if m.UserID == requesterUserID && m.Role == domain.RoleOwner {
			isOwner = true
			break
		}
	}
	if !isOwner {
		return errors.New("only the room owner can remove members")
	}

	// ตรวจสอบว่า member ที่จะลบอยู่ใน room นี้จริง
	target, err := s.memberRepo.GetByID(ctx, roomMemberID)
	if err != nil {
		return errors.New("member not found")
	}
	if target.RoomID != roomID {
		return errors.New("member does not belong to this room")
	}
	if target.Role == domain.RoleOwner {
		return errors.New("cannot remove the room owner")
	}

	return s.memberRepo.DeleteMember(ctx, roomMemberID)
}

func (s *roomMemberService) AddMember(ctx context.Context, roomID, userID uint) (*domain.RoomMember, error) {
	exists, err := s.memberRepo.ExistsByRoomAndUser(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("user is already a member of this room")
	}

	member := &domain.RoomMember{
		RoomID: roomID,
		UserID: userID,
		Role:   domain.RoleMember,
	}
	return s.memberRepo.AddMember(ctx, member)
}

func (s *roomMemberService) CreateInviteCode(ctx context.Context, roomID, creatorUserID uint, access string, expireTime time.Time) (*domain.RoomInviteCode, error) {
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

func (s *roomMemberService) JoinByInviteCode(ctx context.Context, userID uint, inviteCode string) (*domain.RoomMember, error) {
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

func (s *roomMemberService) ListInviteCodes(ctx context.Context, roomID, requesterUserID uint) ([]domain.RoomInviteCode, error) {
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

func (s *roomMemberService) ensureRoomOwner(ctx context.Context, roomID, requesterUserID uint) error {
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

func (s *roomMemberService) generateUniqueInviteCode(ctx context.Context) (string, error) {
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
