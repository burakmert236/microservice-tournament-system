# **GOOD SWIPE**

A distributed, event-driven Go microservices system for managing tournaments, leaderboards, and user rewards.
The platform uses **gRPC**, **DynamoDB single-table design**, **Redis Sorted Sets**, and **event-based communication** to achieve scalability and correctness.

---

# **Table of Contents**

- [**GOOD SWIPE**](#good-swipe)
- [**Table of Contents**](#table-of-contents)
- [**Architecture Overview**](#architecture-overview)
    - [Technologies](#technologies)
- [**Services**](#services)
    - [**1. User Service**](#1-user-service)
    - [**2. Tournament Service**](#2-tournament-service)
    - [**3. Leaderboard Service**](#3-leaderboard-service)
- [**Data Storage**](#data-storage)
    - [**DynamoDB Single-Table Layout**](#dynamodb-single-table-layout)
    - [**Redis**](#redis)
- [**Core Architectural Concepts**](#core-architectural-concepts)
  - [**1. Saga Pattern for Tournament Entry**](#1-saga-pattern-for-tournament-entry)
  - [**2. Idempotency for Reward Claiming**](#2-idempotency-for-reward-claiming)
  - [**3. Event-Based Architecture**](#3-event-based-architecture)
  - [**4. Redis Sorted Lists for Leaderboards**](#4-redis-sorted-lists-for-leaderboards)
- [**Running Locally**](#running-locally)
  - [**Docker Compose**](#docker-compose)
  - [**start.sh**](#startsh)
- [**Environment Variables**](#environment-variables)
- [**Development Workflow**](#development-workflow)

---

# **Architecture Overview**

The system is composed of three main microservices:

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
        │                    │ gRPC │                    │
        │ - User profiles    │      │ - Tournament mgmt  │
        │ - Coin management  │      │ - Group management │
        │ - Reservations     │      │ - Score tracking   │
        │ - Level tracking   │      │ - Reward dist.     │
        └────────────────────┘      └────────────────────┘
                │                              │
                │                              │
                ▼                              ▼
        ┌───────────────────────────────────────────────┐
        │              NATS JetStream                   │
        │         (Event Bus - Port: 4222)              │
        │  ┌──────────────────────────────────────┐     │
        │  │ Events (Protobuf):                   │     │
        │  │ - events.user.created                │     │
        │  │ - events.user.levelup                │     │
        │  │ - events.tournament.entered          │     │
        │  │ - events.tournament.scoreUpdated     │     │
        │  └──────────────────────────────────────┘     │
        └───────────────────────────────────────────────┘
                                  │
                ┌─────────────────┼────────────────────┐
                ▼                 ▼                    ▼
        ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐
        │  DynamoDB    │  │ Leaderboard  │  │    Scheduler         │
        │ (Port: 8000) │  │   Service    │  │  (Cron Jobs)         │
        │              │  │ + Redis      │  │                      │
        │ Single Table │  │ (Port: 9093) │  │ - Create tournaments │
        │   Design     │  │              │  │ - Complete old       │
        └──────────────┘  └──────────────┘  └──────────────────────┘
                                  │
                                  ▼
                          ┌──────────────┐
                          │    Redis     │
                          │ (Port: 6379) │
                          │              │
                          │ Leaderboards │
                          └──────────────┘
```

### Technologies

* **Go + gRPC** microservices
* **DynamoDB** single-table design (user, tournament, participation, reward records)
* **Redis Sorted Sets** for real-time leaderboards
* **NATS JetStream** for asynchronous flows
* **Idempotent service boundaries** to tolerate retries

---

# **Services**

### **1. User Service**

Handles:

* User account state
* Coin balances
* Reservation for saga pattern

Ports:

* **gRPC:** `9091`

### **2. Tournament Service**

Handles:

* Tournament creation
* User participation
* Reward calculation
* Reward claiming (idempotent)

Ports:

* **gRPC:** `9092`

### **3. Leaderboard Service**

Handles:

* Ranking via Redis Sorted Sets
* Update cache by listening events
* Sync with tournament results

Ports:

* **gRPC:** `9093`

---

# **Data Storage**

### **DynamoDB Single-Table Layout**

| Partition Key (PK) | Sort Key (SK)    | Type                                    |
| ------------------ | ---------------- | --------------------------------------- |
| USER#id            | PROFILE             | user profile                            |
| TOURNAMENT#id            | META | tournament meta data               |
| TOURNAMENT#id           | GROUP#id             | tournament group                     |
| USER#id           | TORUNAMENT#id      | participation                           |
| RESERVATION#id           | META             | reservation for tournament entry |
| REWARDCLAIM#id           | TORUNAMENT#id             | idempotency tracking for tournamnt reward |

### **Redis**

Used for:

* Sorted sets to maintain tournament rankings
* O( log N ) updates
* O( N ) retrieval for top N

Example key:

```
leaderboard:group:{tournamentId}:{groupId}
```

---

# **Core Architectural Concepts**

## **1. Saga Pattern for Tournament Entry**

Tournament entry includes multiple steps:

1. Validate user
2. Create participation record
3. Publish "user joined" event
4. Update leaderboard

Failures at any step require compensation:

* Roll back participation
* Remove leaderboard scores
* Emit compensating event

This guarantees **eventual consistency** across services without distributed transactions.

---

## **2. Idempotency for Reward Claiming**

Reward claims must be safe to retry (network failures, UI retries, async processing).

Mechanism:

* Participation row contains `claimStatus = NONE | LOCKED | CLAIMED`
* Update uses conditional write:
  `SET claimStatus = LOCKED IF claimStatus = NONE`
* Reward credit uses **reasonType + reasonId** (claimId) at User Service
* User ledger ensures **each reward is credited exactly once**

This prevents:

* Double-rewarding
* Race conditions
* Inconsistent state

---

## **3. Event-Based Architecture**

Services communicate by emitting domain events, e.g.:

* `UserJoinedTournament`
* `ScoreUpdated`
* `RewardClaimed`

Benefits:

* Loose coupling
* Replayability
* Easy to add analytics / notifications

---

## **4. Redis Sorted Lists for Leaderboards**

Scores are written as:

```
ZADD LEADERBOARD:<tournamentId> score userId
```

Queries:

* Top N: `ZREVRANGE key 0 N WITHSCORES`
* User rank: `ZREVRANK key userId`

This makes the leaderboard service extremely fast and scalable.

---

# **Running Locally**

## **Docker Compose**




## **start.sh**


---

# **Environment Variables**

| Variable                   | Purpose                        |
| -------------------------- | ------------------------------ |
| `DYNAMODB_ENDPOINT`        | Local or AWS DynamoDB endpoint |
| `REDIS_HOST`               | Redis location for leaderboard |
| `USER_SERVICE_ADDR`        | gRPC address                   |
| `TOURNAMENT_SERVICE_ADDR`  | gRPC address                   |
| `LEADERBOARD_SERVICE_ADDR` | gRPC address                   |

---

# **Development Workflow**

1. `./start.sh`
2. Use `grpcurl` to test endpoints
3. Modify protobuf → regenerate code via `buf` or `protoc`
4. Run individual services locally for debugging:

```bash
go run cmd/tournament/main.go
```

---
