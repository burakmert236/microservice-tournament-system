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
        │   Design     │  │              │  │                      │
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
2. Create tournament reservation via User Service
3. Deduct enterance fee from user
4. Create participation and update group in transaction
5. Confirm reservation via User Service
6. Publish "user joined" event

Transaction failure require compensation:

* Roll back reservation
* Return enterance fee to user

This guarantees **eventual consistency** across services without distributed transactions.

---

## **2. Idempotency for Reward Claiming**

Reward claims must be safe to retry (network failures, UI retries, async processing).

Mechanism:

* Participation row contains `reward_claim_status = UNCLAIMED | PROCESSING | CLAIMED`
* Update uses conditional write:
  `SET reward_claim_status = PROCESSING IF reward_claim_status = UNCLAIMED`
* Get ranking data from Leaderboard service
* User service gets request for reward 
* User service stores reward claims as userId + tournamentId 
* Reward claim entries checked for prevent double rewarding
* User reward claims ensures **each reward is credited exactly once**

This prevents:

* Double-rewarding
* Race conditions
* Inconsistent state

---

## **3. Event-Based Architecture**

Services communicate by emitting domain events, e.g.:

* `UserCreated`
* `UserLevelUp`
* `TournamentEntered`
* `TournamentParticipationScoreUpdated`

Benefits:

* Loose coupling
* Replayability
* Easy to add analytics / notifications

---

## **4. Redis Sorted Lists for Leaderboards**

Scores are written as:

```
ZADD leaderboard:group:{tournamentId}:{groupId} score userId
```

This makes the leaderboard service extremely fast and scalable.

---

# **Running Locally**

## **Docker Compose**

Docker compose file includes every necessary service and running compose file is sufficient to start entire system:

```
docker compose up --build -d
```

## **start.sh**

A trivial start script. It contains only docker compose command.

```
chmod +x start.sh
./start.sh
```
