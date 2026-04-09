# Frontend Agent Docs: Room, Members, Invite Codes

## 1. Scope
เอกสารนี้อธิบาย API และกติกาธุรกิจสำหรับฟีเจอร์:
- Room (เฉพาะข้อมูลที่เกี่ยวกับ flow ปัจจุบัน)
- Room Members
- Room Invite Codes

เอกสารนี้ตั้งใจให้ frontend agent ใช้เป็น source of truth ตอนสร้างหน้า UI, service, hooks, และ error handling

## 2. Domain Summary (Current Backend)
- User
- Room
- RoomMember
- RoomInviteCode
- Trip (ใช้สร้างห้องใน flow เริ่มต้น)

ความสัมพันธ์สำคัญ:
- 1 Room มีหลาย RoomMember
- RoomMember เก็บ role: owner หรือ member
- RoomInviteCode ผูกกับ Room และมีผู้สร้างโค้ด
- Invite code มี access (view/edit) และ expire_time

หมายเหตุ:
- ปัจจุบันยังไม่มี endpoint CRUD ห้องโดยตรง
- ห้องถูกสร้างผ่าน flow สร้างทริป (Create Trip)

## 3. Base API Behavior
### 3.1 Response Envelope
ทุก endpoint คืนโครงแบบเดียวกัน:

```json
{
  "status": 200,
  "message": "success",
  "data": {},
  "error": ""
}
```

- data จะไม่มีเมื่อเกิด error
- error จะมีข้อความเมื่อ request ล้มเหลว

### 3.2 Auth
ทุก endpoint ในกลุ่ม room member / invite code ต้องส่ง header:

```http
Authorization: Bearer <token>
```

ถ้า token ไม่ถูกต้องจะได้ 401 พร้อม message = unauthorized

### 3.3 Quick Start API Examples

```bash
# 1) My rooms
curl -X GET "$BASE_URL/api/rooms/user/123" \
  -H "Authorization: Bearer $TOKEN"

# 2) Room members
curl -X GET "$BASE_URL/api/rooms/10/members" \
  -H "Authorization: Bearer $TOKEN"

# 3) Create invite code
curl -X POST "$BASE_URL/api/rooms/10/invite-codes" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"access":"view","expire_time":"2026-04-10T12:30:00Z"}'

# 4) Join by invite code
curl -X POST "$BASE_URL/api/rooms/join-by-invite-code" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"invite_code":"ABCD23EF"}'
```

## 4. Endpoint Map (Room + Members + Invite Codes)

### 4.1 Members
1. GET /api/rooms/user/:userID
2. GET /api/rooms/:roomID/members
3. POST /api/rooms/:roomID/members
4. DELETE /api/rooms/:roomID/members/:memberID

### 4.2 Invite Codes
1. GET /api/rooms/:roomID/invite-codes
2. POST /api/rooms/:roomID/invite-codes
3. POST /api/rooms/join-by-invite-code

## 5. Contracts Per Endpoint

## 5.1 GET /api/rooms/user/:userID
Auth: Required

Rules:
- ผู้ใช้เรียกได้เฉพาะ userID ของตัวเอง (ถ้าเรียก user อื่นจะ 403)

Success data shape:

```json
[
  {
    "room_id": 10,
    "room_name": "Trip to Bangkok",
    "room_image": "https://...",
    "owner_id": 99,
    "owner_username": "alice",
    "role": 2,
    "role_name": "member",
    "joined_at": "2026-04-09T12:45:00Z",
    "members_count": 3
  }
]
```

Common errors:
- 400: userID must be a number
- 403: cannot access other user's rooms
- 500: failed to get user rooms

## 5.2 GET /api/rooms/:roomID/members
Auth: Required

Success data shape:

```json
[
  {
    "room_member_id": 1,
    "room_id": 10,
    "user_id": 99,
    "username": "alice",
    "role": 1,
    "role_name": "owner",
    "created_at": "2026-04-09T10:00:00+07:00"
  }
]
```

Common errors:
- 400: roomID must be a number
- 500: failed to get room members

## 5.3 POST /api/rooms/:roomID/members
Auth: Required

Request:

```json
{
  "user_id": 123
}
```

Success 201 data shape:

```json
{
  "room_member_id": 2,
  "room_id": 10,
  "user_id": 123,
  "username": "bob",
  "role": 2,
  "role_name": "member",
  "created_at": "2026-04-09T10:05:00+07:00"
}
```

Common errors:
- 400: user_id is required
- 400: user is already a member of this room

## 5.4 DELETE /api/rooms/:roomID/members/:memberID
Auth: Required

Success 200 data: none

Common errors:
- 400: only the room owner can remove members
- 400: member does not belong to this room
- 400: cannot remove the room owner

## 5.5 POST /api/rooms/:roomID/invite-codes
Auth: Required (ต้องเป็น owner)

Request:

```json
{
  "access": "view",
  "expire_time": "2026-04-10T12:30:00Z"
}
```

Rules:
- access รองรับ view หรือ edit
- ถ้าไม่ส่ง access จะ default เป็น view
- expire_time รองรับ 2 format:
  - RFC3339 (แนะนำ)
  - YYYY-MM-DD HH:MM:SS
- ถ้าไม่ส่ง expire_time จะ default = now + 24h

Success 201 data shape:

```json
{
  "room_invite_id": 1,
  "room_id": 10,
  "invite_code_creator_id": 99,
  "invite_code": "ABCD23EF",
  "access": "view",
  "expire_time": "2026-04-10T12:30:00Z",
  "created_at": "2026-04-09T12:30:00Z"
}
```

Common errors:
- 400: access must be view or edit
- 400: expire_time must be in the future
- 400: only the room owner can perform this action

## 5.6 GET /api/rooms/:roomID/invite-codes
Auth: Required (ต้องเป็น owner)

Success 200 data shape:

```json
[
  {
    "room_invite_id": 1,
    "room_id": 10,
    "invite_code_creator_id": 99,
    "invite_code": "ABCD23EF",
    "access": "view",
    "expire_time": "2026-04-10T12:30:00Z",
    "created_at": "2026-04-09T12:30:00Z"
  }
]
```

หมายเหตุ:
- backend คืนเฉพาะโค้ดที่ยังไม่หมดอายุ

Common errors:
- 400: only the room owner can perform this action

## 5.7 POST /api/rooms/join-by-invite-code
Auth: Required

Request:

```json
{
  "invite_code": "ABCD23EF"
}
```

Success 201 data shape (RoomMemberResponseDTO):

```json
{
  "room_member_id": 3,
  "room_id": 10,
  "user_id": 555,
  "username": "charlie",
  "role": 2,
  "role_name": "member",
  "created_at": "2026-04-09T12:45:00+07:00"
}
```

Common errors:
- 400: invite_code is required
- 400: invite code not found
- 400: invite code has expired
- 400: user is already a member of this room

## 6. Frontend Integration Notes

## 6.1 Suggested TypeScript Types

```ts
export type ApiResponse<T> = {
  status: number;
  message: string;
  data?: T;
  error?: string;
};

export type RoomMember = {
  room_member_id: number;
  room_id: number;
  user_id: number;
  username: string;
  role: number;
  role_name: "owner" | "member" | "unknown";
  created_at: string;
};

export type UserRoomSummary = {
  room_id: number;
  room_name: string;
  room_image: string;
  owner_id: number;
  owner_username: string;
  role: number;
  role_name: "owner" | "member" | "unknown";
  joined_at: string;
  members_count: number;
};

export type RoomInviteCode = {
  room_invite_id: number;
  room_id: number;
  invite_code_creator_id: number;
  invite_code: string;
  access: "view" | "edit";
  expire_time: string;
  created_at: string;
};
``