# Good Swipe

A high-performance, scalable microservices architecture for a competitive Match-3 game tournament system built with Go, gRPC, DynamoDB, Redis, and NATS JetStream.

---

## 📋 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Technology Stack](#technology-stack)
- [Key Concepts](#key-concepts)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [API Endpoints](#api-endpoints)
- [Design Patterns](#design-patterns)
- [Data Models](#data-models)
- [Testing](#testing)
- [Performance & Scalability](#performance--scalability)
- [Monitoring](#monitoring)

---

## 🎯 Overview

Good Swipe Tournament System enables millions of players to compete in daily tournaments with the following features:

### **Core Features**
- ✅ **User Management**: Track levels, coins, and progression
- ✅ **Daily Tournaments**: Automated creation at 00:00 UTC, ending at 23:59 UTC
- ✅ **Tournament Entry**: Level-gated (≥10), coin-based entry fee (500 coins)
- ✅ **Group-Based Competition**: 35 players per group
- ✅ **Real-time Leaderboards**: Global and tournament-specific rankings
- ✅ **Reward Distribution**: Automated ranking-based rewards
- ✅ **Event-Driven Architecture**: Asynchronous score updates via NATS JetStream

### **Technical Highlights**
- 🚀 **High Performance**: Sub-millisecond response times
- 🔒 **Data Consistency**: ACID transactions with DynamoDB
- 📈 **Horizontally Scalable**: Stateless microservices architecture
- 🔄 **Idempotent Operations**: Safe retries and duplicate request handling
- 🎪 **Saga Pattern**: Distributed transaction management with compensation
- 📊 **Event Sourcing**: Score updates via protobuf events

---

## 🏗️ Architecture

### **High-Level Architecture**

```
┌─────────────────────────────────────────────────────────────────────┐
│                          Client Layer                               │
│                      (gRPC / Postman)                               │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                    ┌─────────────┴─────────────┐
                    ▼                           ▼
        ┌────────────────────┐      ┌────────────────────┐
        │   User Service     │      │ Tournament Service │
        │   (Port: 9091)     │◄────►│   (Port: 9092)     │
        └────────────────────┘      └────────────────────┘
                │                              │
                │                              │
                ▼                              ▼
        ┌────────────────────────────────────────────────┐
        │              NATS JetStream                    │
        │         (Event Bus - Port: 4222)               │
        │  ┌──────────────────────────────────────┐      │
        │  │ Subjects:                            │      │
        │  │ - events.user.levelup                │      │
        │  │ - events.tournament.completed        │      │
        │  └──────────────────────────────────────┘      │
        └────────────────────────────────────────────────┘
                                  │
                ┌─────────────────┼─────────────────┐
                ▼                 ▼                 ▼
        ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
        │  DynamoDB    │  │    Redis     │  │  Scheduler   │
        │ (Port: 8000) │  │ (Port: 6379) │  │ (Cron Jobs)  │
        └──────────────┘  └──────────────┘  └──────────────┘
```

### **Service Responsibilities**

#### **User Service**
- User creation and profile management
- Level progression tracking
- Coin balance management
- Reservation system (for tournament entry)
- Event publishing (user level-up events)

#### **Tournament Service**
- Daily tournament creation (scheduled at 00:00 UTC)
- Tournament entry orchestration (Saga pattern)
- Group management (35 players per group)
- Participation tracking
- Reward distribution
- Score updates via event consumption

#### **Scheduler**
- Automated tournament creation (00:00 UTC daily)
- Tournament completion and finalization (23:59 UTC)
- Expired reservation cleanup

---

## 🛠️ Technology Stack

### **Backend**
- **Language**: Go 1.21+
- **RPC Framework**: gRPC with Protocol Buffers
- **Primary Database**: AWS DynamoDB (Single Table Design)
- **Cache Layer**: Redis (Leaderboards)
- **Message Broker**: NATS JetStream (Event streaming)
- **Containerization**: Docker & Docker Compose

### **Key Libraries**
```go
- google.golang.org/grpc          // gRPC server/client
- google.golang.org/protobuf      // Protocol Buffers
- github.com/aws/aws-sdk-go-v2    // AWS SDK for DynamoDB
- github.com/nats-io/nats.go      // NATS client
- github.com/redis/go-redis/v9    // Redis client
- github.com/spf13/viper          // Configuration management
- github.com/google/uuid          // UUID generation
```

### **Development Tools**
- Docker & Docker Compose
- Postman (API testing)
- AWS CLI (DynamoDB management)

---

## 💡 Key Concepts

### **1. Saga Pattern (Distributed Transactions)**

The tournament entry process requires coordination between User and Tournament services. We use the **Saga Pattern with Compensation** for reliable distributed transactions:

```
EnterTournament Flow (Saga):

1. Tournament Service validates entry requirements
   ├─ Level >= 10
   ├─ Time < 12:00 UTC
   └─ Not already participating

2. [gRPC] Tournament → User: ReserveCoins(500)
   └─ User Service: Deduct 500 coins + Create reservation record
      (Single DynamoDB transaction)

3. Tournament Service: Add user to group
   └─ Create participant + Increment group count
      (Single DynamoDB transaction)

4. Success Path:
   └─ [gRPC] Tournament → User: ConfirmReservation()
      └─ Mark reservation as CONFIRMED

5. Failure Path (Compensation):
   └─ [gRPC] Tournament → User: RollbackReservation()
      └─ Refund 500 coins + Mark as ROLLED_BACK
```

**Why Saga Pattern?**
- ✅ No distributed locks required
- ✅ Each service maintains its own transaction boundary
- ✅ Automatic compensation on failure
- ✅ Idempotent operations with reservation IDs

### **2. Event-Driven Architecture**

Asynchronous operations use NATS JetStream for reliable event delivery:

```
User Level-Up Flow (Event-Driven):

1. User Service: UpdateProgress(userID, levelIncrease)
   └─ Update user level + coins (DynamoDB transaction)

2. User Service: Publish event to NATS
   └─ Subject: events.user.levelup
   └─ Payload: UserLeveledUpEvent (protobuf)
      {
        eventId: "uuid-123",
        userId: "user-456",
        newLevel: 15,
        oldLevel: 14,
        timestamp: "2024-01-15T10:30:00Z"
      }

3. Tournament Service: Consume event from NATS
   └─ Check if user has active participations
   └─ Calculate score bonus (newLevel * 10)
   └─ Update participation score (ADD operation)
      (Atomic DynamoDB update)
```

**Why Event-Driven?**
- ✅ Loose coupling between services
- ✅ Asynchronous processing (non-blocking)
- ✅ Scalable (multiple consumers)
- ✅ Reliable delivery (NATS persistence)
- ✅ Event replay capability

### **3. Eventual Consistency**

The system balances **strong consistency** (where needed) with **eventual consistency** (where acceptable):

#### **Strong Consistency** (Immediate)
- ✅ Coin deductions (tournament entry)
- ✅ User balance updates
- ✅ Group participant count
- ✅ Reservation state changes

#### **Eventual Consistency** (Seconds delay acceptable)
- ⏱️ Tournament score updates (from level-ups)
- ⏱️ Leaderboard rankings (5-30 second cache)
- ⏱️ Analytics and metrics

**Trade-off**: Prioritizes performance and scalability while maintaining critical data correctness.

### **4. Idempotency**

All critical operations are designed to be **idempotent** - calling them multiple times with the same parameters produces the same result:

#### **Idempotency Strategies**

**A. Reservation IDs (Tournament Entry)**
```go
// Client generates UUID
reservationID := uuid.New().String()

// All operations use this ID
ReserveCoins(userID, 500, reservationID)
AddToGroup(userID, groupID, reservationID)
ConfirmReservation(reservationID)

// Retry with same ID → No duplicate charges
```

**B. Event IDs (Score Updates)**
```go
event := UserLeveledUpEvent{
    EventId: "uuid-123",  // Unique event identifier
    UserId: "user-456",
    // ...
}

// Consumer checks if already processed
if processedEvents[event.EventId] {
    return // Skip duplicate
}

processScoreUpdate(event)
processedEvents[event.EventId] = true
```

**C. Conditional Expressions (DynamoDB)**
```go
// Only create if doesn't exist
ConditionExpression: "attribute_not_exists(PK)"

// Only update if not already completed
ConditionExpression: "attribute_not_exists(completedAt)"
```

### **5. Single Table Design (DynamoDB)**

All entities stored in one DynamoDB table with strategic key design:

```
Table: GameTable

Primary Key: PK (Partition Key) + SK (Sort Key)
GSI1: GSI1PK (Partition Key) + GSI1SK (Sort Key)

Entity Examples:
┌─────────────────────────────────┬──────────────┬────────────────┐
│ Entity                          │ PK           │ SK             │
├─────────────────────────────────┼──────────────┼────────────────┤
│ User Profile                    │ USER#123     │ PROFILE        │
│ Tournament Metadata             │ TOURNAMENT#1 │ META           │
│ Tournament Group                │ TOURNAMENT#1 │ GROUP#G1       │
│ Participant (in group)          │ TOURNEY#1#G1 │ USER#123       │
│ Reservation                     │ RESERVATION#R│ META           │
└─────────────────────────────────┴──────────────┴────────────────┘

GSI1 Examples (Alternative access patterns):
┌─────────────────────────────────┬──────────────────┬─────────────┐
│ Access Pattern                  │ GSI1PK           │ GSI1SK      │
├─────────────────────────────────┼──────────────────┼─────────────┤
│ Current tournaments             │ CURRENT_TOURN    │ START#time  │
│ User's tournaments history      │ USER#123         │ TOURN#1#... │
└─────────────────────────────────┴──────────────────┴─────────────┘
```

**Benefits:**
- ✅ Single table = simpler operations
- ✅ Lower cost (no cross-table queries)
- ✅ Better performance (co-located data)
- ✅ Atomic transactions within partition

---

## 📁 Project Structure

```
good-swipe/
├── common/                          # Shared code across services
│   ├── config/                      # Configuration management
│   │   ├── config.go               # Config structs and loaders
│   │   └── env.go                  # Environment variable handling
│   ├── db/                         # Database clients
│   │   ├── dynamodb.go             # DynamoDB client setup
│   │   └── transaction.go          # Transaction helpers
│   ├── messaging/                  # NATS client
│   │   └── nats_client.go
│   ├── models/                     # Domain models
│   │   ├── user.go
│   │   ├── tournament.go
│   │   ├── reservation.go
│   │   └── keys.go                 # DynamoDB key builders
│   ├── proto/                      # Protocol Buffer definitions
│   │   ├── user/
│   │   │   ├── user.proto
│   │   │   ├── user.pb.go
│   │   │   └── user_grpc.pb.go
│   │   ├── tournament/
│   │   │   ├── tournament.proto
│   │   │   ├── tournament.pb.go
│   │   │   └── tournament_grpc.pb.go
│   │   └── events/
│   │       ├── user_events.proto
│   │       └── user_events.pb.go
│   ├── errors/                     # Custom error types
│   │   └── errors.go
│   └── middleware/                 # gRPC middleware
│       ├── logging.go
│       └── recovery.go
│
├── services/
│   ├── user-service/               # User management service
│   │   ├── cmd/
│   │   │   └── main.go             # Service entry point
│   │   ├── internal/
│   │   │   ├── handler/            # gRPC handlers
│   │   │   │   └── user_handler.go
│   │   │   ├── service/            # Business logic
│   │   │   │   └── user_service.go
│   │   │   ├── repository/         # Data access layer
│   │   │   │   ├── user_repo.go
│   │   │   │   └── reservation_repo.go
│   │   │   └── messaging/          # Event publishing
│   │   │       └── event_publisher.go
│   │   ├── config/
│   │   │   └── config.yaml
│   │   ├── Dockerfile
│   │   └── go.mod
│   │
│   └── tournament-service/         # Tournament management service
│       ├── cmd/
│       │   └── main.go
│       ├── internal/
│       │   ├── handler/            # gRPC handlers
│       │   │   └── tournament_handler.go
│       │   ├── service/            # Business logic
│       │   │   ├── tournament_service.go
│       │   │   └── participant_service.go
│       │   ├── repository/         # Data access layer
│       │   │   ├── tournament_repo.go
│       │   │   ├── group_repo.go
│       │   │   └── participant_repo.go
│       │   ├── scheduler/          # Scheduled tasks
│       │   │   ├── scheduler.go
│       │   │   └── tournament_scheduler.go
│       │   └── messaging/          # Event consumption
│       │       └── event_consumer.go
│       ├── config/
│       │   └── config.yaml
│       ├── Dockerfile
│       └── go.mod
│
├── scripts/                        # Utility scripts
│   ├── init-dynamodb.sh           # DynamoDB table initialization
│   └── generate-proto.sh          # Protobuf code generation
│
├── schemas/                        # Database schemas
│   └── dynamodb-table.json        # DynamoDB table definition
│
├── docker-compose.yaml            # Local development environment
├── Makefile                       # Build and run commands
├── postman_collection.json        # API test collection
└── README.md                      # This file
```

---

## 🚀 Getting Started

### **Prerequisites**

- Docker & Docker Compose (v20.10+)
- Go 1.21+ (for local development)
- Make (optional, for convenience)

### **Quick Start (Docker Compose)**

#### **1. Clone the Repository**

```bash
git clone <repository-url>
cd good-swipe
```

#### **2. Start All Services**

```bash
# Start entire system with one command
docker-compose up --build

# Or use detached mode
docker-compose up -d --build
```

This will start:
- ✅ DynamoDB Local (port 8000)
- ✅ NATS JetStream (port 4222)
- ✅ Redis (port 6379)
- ✅ User Service (port 9091)
- ✅ Tournament Service (port 9092)

#### **3. Verify Services are Running**

```bash
# Check service status
docker-compose ps

# Expected output:
NAME                  STATUS    PORTS
dynamodb             Up        0.0.0.0:8000->8000/tcp
nats                 Up        0.0.0.0:4222->4222/tcp
redis                Up        0.0.0.0:6379->6379/tcp
user-service         Up        0.0.0.0:9091->9091/tcp
tournament-service   Up        0.0.0.0:9092->9092/tcp
```

#### **4. View Logs**

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f user-service
docker-compose logs -f tournament-service
```

#### **5. Test the APIs**

Import `postman_collection.json` into Postman or use grpcurl:

```bash
# Create a user
grpcurl -plaintext -d '{
  "display_name": "Player1"
}' localhost:9091 user.UserService/CreateUser

# Response:
# {
#   "userId": "550e8400-e29b-41d4-a716-446655440000",
#   "displayName": "Player1",
#   "level": 1,
#   "coin": 1000
# }
```

#### **6. Stop Services**

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (clean state)
docker-compose down -v
```

---

## 📡 API Endpoints

### **User Service (Port 9091)**

#### **CreateUser**
```protobuf
rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);

Request:
{
  "display_name": "Player1"
}

Response:
{
  "user_id": "uuid",
  "display_name": "Player1",
  "level": 1,
  "coin": 1000
}
```

#### **UpdateProgress**
```protobuf
rpc UpdateProgress(UpdateProgressRequest) returns (UpdateProgressResponse);

Request:
{
  "user_id": "uuid",
  "level_increase": 2
}

Response:
{
  "success": true,
  "new_level": 3,
  "coins_earned": 200
}

Side Effects:
- Publishes UserLeveledUpEvent to NATS
- Tournament scores updated asynchronously
```

#### **GetUser**
```protobuf
rpc GetUser(GetUserRequest) returns (GetUserResponse);

Request:
{
  "user_id": "uuid"
}

Response:
{
  "user": {
    "user_id": "uuid",
    "display_name": "Player1",
    "level": 10,
    "coin": 1500
  }
}
```

### **Tournament Service (Port 9092)**

#### **EnterTournament** (Idempotent)
```protobuf
rpc EnterTournament(EnterTournamentRequest) returns (EnterTournamentResponse);

Request:
{
  "user_id": "uuid"
}

Response (Success):
{
  "success": true,
  "group_id": "group-uuid",
  "message": "Successfully joined tournament"
}

Response (Failure):
{
  "success": false,
  "message": "insufficient coins" | "level requirement not met" | "entry closed"
}

Requirements:
- User level >= 10
- User has >= 500 coins
- Current time < 12:00 UTC
- Not already participating

Process (Saga Pattern):
1. Validate entry requirements
2. Reserve 500 coins (User Service)
3. Add to tournament group
4. Confirm reservation OR rollback on failure
```

#### **ClaimReward** (Idempotent)
```protobuf
rpc ClaimReward(ClaimRewardRequest) returns (ClaimRewardResponse);

Request:
{
  "user_id": "uuid"
}

Response:
{
  "success": true,
  "rank": 1,
  "coins_earned": 5000
}

Rewards:
- 1st place: 5000 coins
- 2nd place: 3000 coins
- 3rd place: 2000 coins
- 4th-10th: 1000 coins each
```

#### **GetTournamentLeaderboard**
```protobuf
rpc GetTournamentLeaderboard(GetTournamentLeaderboardRequest) returns (GetTournamentLeaderboardResponse);

Request:
{
  "user_id": "uuid"
}

Response:
{
  "leaderboard": [
    {
      "user_id": "uuid",
      "display_name": "Player1",
      "score": 150,
      "rank": 1
    },
    // ... up to 35 entries (group size)
  ]
}

Cache: 5 seconds freshness
```

#### **GetTournamentRank**
```protobuf
rpc GetTournamentRank(GetTournamentRankRequest) returns (GetTournamentRankResponse);

Request:
{
  "user_id": "uuid"
}

Response:
{
  "rank": 5,
  "score": 120,
  "total_participants": 35
}
```

### **GetGlobalLeaderboard**
```protobuf
rpc GetGlobalLeaderboard(GetGlobalLeaderboardRequest) returns (GetGlobalLeaderboardResponse);

Response:
{
  "leaderboard": [
    {
      "user_id": "uuid",
      "display_name": "TopPlayer",
      "level": 99,
      "score": 9900
    },
    // ... top 1000 users
  ]
}

Cache: 30 seconds freshness
```

---

## 🎨 Design Patterns

### **1. Repository Pattern**
Abstracts data access logic from business logic:

```go
// Interface (abstraction)
type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    GetByID(ctx context.Context, userID string) (*models.User, error)
    UpdateLevelAndCoins(ctx context.Context, userID string, level int, coins int64) error
}

// Implementation (concrete)
type userRepo struct {
    db *db.DynamoDBClient
}

// Business logic depends on interface, not implementation
type UserService struct {
    userRepo UserRepository  // ← Interface
}
```

### **2. Dependency Injection**
Services receive dependencies through constructors:

```go
// services/tournament-service/cmd/main.go
func main() {
    // Initialize dependencies
    dynamoClient := db.NewDynamoDBClient(cfg)
    userClient := userpb.NewUserServiceClient(conn)
    
    // Inject dependencies
    repo := repository.NewTournamentRepository(dynamoClient)
    service := service.NewTournamentService(repo, userClient)
    handler := handler.NewTournamentHandler(service)
    
    // Register handler
    pb.RegisterTournamentServiceServer(grpcServer, handler)
}
```

### **3. Saga Pattern (Orchestration)**
Tournament Service orchestrates the entry flow:

```go
func (s *tournamentService) EnterTournament(ctx context.Context, userID string) error {
    reservationID := uuid.New().String()
    
    // Step 1: Reserve coins
    err := s.userClient.ReserveCoins(ctx, userID, 500, reservationID)
    if err != nil {
        return err
    }
    
    // Step 2: Add to group
    err = s.addToGroup(ctx, userID)
    if err != nil {
        // Compensation: Rollback
        s.userClient.RollbackReservation(ctx, reservationID)
        return err
    }
    
    // Step 3: Confirm
    s.userClient.ConfirmReservation(ctx, reservationID)
    return nil
}
```

### **4. Event Sourcing (Lite)**
State changes published as events:

```go
// User Service
func (s *userService) UpdateProgress(ctx context.Context, userID string, levelInc int) error {
    // Update state
    newLevel, err := s.repo.IncrementLevel(ctx, userID, levelInc)
    if err != nil {
        return err
    }
    
    // Publish event
    event := &UserLeveledUpEvent{
        EventId: uuid.New().String(),
        UserId: userID,
        NewLevel: newLevel,
    }
    s.eventPublisher.Publish(ctx, event)
    
    return nil
}
```

### **5. Circuit Breaker (Future Enhancement)**
Prevents cascading failures:

```go
// If User Service is down, don't keep trying
if circuitBreaker.IsOpen() {
    return errors.New("user service unavailable")
}

err := s.userClient.ReserveCoins(ctx, ...)
if err != nil {
    circuitBreaker.RecordFailure()
}
```

---

## 📊 Data Models

### **User**
```go
type User struct {
    UserID      string    `dynamodbav:"userId"`
    DisplayName string    `dynamodbav:"displayName"`
    Level       int       `dynamodbav:"level"`
    Coin        int64     `dynamodbav:"coin"`
    CreatedAt   time.Time `dynamodbav:"createdAt"`
    UpdatedAt   time.Time `dynamodbav:"updatedAt"`
    
    PK string `dynamodbav:"PK"` // USER#<userId>
    SK string `dynamodbav:"SK"` // PROFILE
}
```

### **Tournament**
```go
type Tournament struct {
    TournamentID string            `dynamodbav:"tournamentId"`
    Name         string            `dynamodbav:"name"`
    GroupSize    int               `dynamodbav:"groupSize"` // 35
    EntranceFee  int64             `dynamodbav:"entranceFee"` // 500
    StartTime    time.Time         `dynamodbav:"startTime"`
    EndTime      time.Time         `dynamodbav:"endTime"`
    Rewards      map[int]int64     `dynamodbav:"rewards"` // rank -> coins
    CreatedAt    time.Time         `dynamodbav:"createdAt"`
    
    PK     string `dynamodbav:"PK"` // TOURNAMENT#<tournamentId>
    SK     string `dynamodbav:"SK"` // META
    GSI1PK string `dynamodbav:"GSI1PK"` // CURRENT_TOURNAMENT
    GSI1SK string `dynamodbav:"GSI1SK"` // START#<timestamp>
}
```

### **Group**
```go
type Group struct {
    GroupID          string    `dynamodbav:"groupId"`
    TournamentID     string    `dynamodbav:"tournamentId"`
    ParticipantCount int       `dynamodbav:"participantCount"`
    GroupSize        int       `dynamodbav:"groupSize"` // 35
    CreatedAt        time.Time `dynamodbav:"createdAt"`
    
    PK string `dynamodbav:"PK"` // TOURNAMENT#<tournamentId>
    SK string `dynamodbav:"SK"` // GROUP#<groupId>
}
```

### **Participant**
```go
type TournamentParticipant struct {
    UserID       string     `dynamodbav:"userId"`
    TournamentID string     `dynamodbav:"tournamentId"`
    GroupID      string     `dynamodbav:"groupId"`
    Score        int64      `dynamodbav:"score"`
    Rank         int        `dynamodbav:"rank"`
    CoinsEarned  int64      `dynamodbav:"coinsEarned"`
    JoinedAt     time.Time  `dynamodbav:"joinedAt"`
    CompletedAt  *time.Time `dynamodbav:"completedAt,omitempty"`
    
    PK string `dynamodbav:"PK"` // TOURNAMENT#<id>#GROUP#<id>
    SK string `dynamodbav:"SK"` // USER#<userId>
    
    GSI1PK string `dynamodbav:"GSI1PK"` // USER#<userId>
    GSI1SK string `dynamodbav:"GSI1SK"` // TOURNAMENT#<id>#JOINED#<time>
}
```

### **Reservation**
```go
type Reservation struct {
    ReservationID string            `dynamodbav:"reservationId"`
    UserID        string            `dynamodbav:"userId"`
    Amount        int64             `dynamodbav:"amount"`
    Status        ReservationStatus `dynamodbav:"status"` // RESERVED, CONFIRMED, ROLLED_BACK
    Purpose       string            `dynamodbav:"purpose"` // TOURNAMENT_ENTRY
    CreatedAt     time.Time         `dynamodbav:"createdAt"`
    ExpiresAt     time.Time         `dynamodbav:"expiresAt"` // 5 minutes
    
    PK string `dynamodbav:"PK"` // RESERVATION#<reservationId>
    SK string `dynamodbav:"SK"` // META
}
```

---

## 🧪 Testing

### **Unit Tests**

```bash
# Test all packages
make test

# Test specific service
cd services/user-service
go test ./...

# Test with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### **Integration Tests**

```bash
# Start test environment
docker-compose -f docker-compose.test.yaml up -d

# Run integration tests
make integration-test

# Cleanup
docker-compose -f docker-compose.test.yaml down -v
```

### **Load Testing**

```bash
# Install k6
brew install k6  # macOS
# or download from https://k6.io/

# Run load test
k6 run scripts/load-test.js

# Example scenario:
# - 1000 virtual users
# - 10,000 requests per second
# - 5 minute duration
```

### **Example Test**

```go
func TestEnterTournament_Idempotency(t *testing.T) {
    // Setup
    service := setupTestService(t)
    userID := createTestUser(t, service, level=15, coins=1000)
    reservationID := uuid.New().String()
    
    // First call - should succeed
    participant1, err := service.Enter