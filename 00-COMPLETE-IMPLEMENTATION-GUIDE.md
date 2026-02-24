# Gym Tracker Microservices - Complete Implementation Guide
## Final Version - Ready to Build

**Project Duration:** Feb 16-22, 2025 (7 days)
**Tech Stack:** Go, PostgreSQL, Redis, Docker, Docker Compose

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Database Schemas](#database-schemas)
3. [API Response Structure](#api-response-structure)
4. [Service Specifications](#service-specifications)
5. [Analytics Data Flow](#analytics-data-flow)
6. [Day-by-Day Implementation Plan](#day-by-day-implementation-plan)
7. [Docker Setup](#docker-setup)
8. [Environment Configuration](#environment-configuration)
9. [Testing Strategy](#testing-strategy)
10. [Quick Start Commands](#quick-start-commands)

---

## Architecture Overview

### Service Diagram

```
                    ┌─────────────┐
                    │   API       │
                    │  Gateway    │ (8080)
                    └──────┬──────┘
                           │
        ┏━━━━━━━━━━━━━━━━━━┻━━━━━━━━━━━━━━━━━━┓
        ┃                                      ┃
┌───────▼────────┐                    ┌────────▼────────┐
│     User       │                    │    Exercise     │
│   Service      │                    │    Service      │
│    (8081)      │                    │     (8084)      │
└───────┬────────┘                    └────────┬────────┘
        │                                      │
        ┃         ┌──────────────┐            ┃
        ┗━━━━━━━━▶│   Workout    │◀━━━━━━━━━━┛
                  │   Service    │
                  │    (8082)    │
                  └──────┬───────┘
                         │
                  ┌──────▼───────┐         ┌────────────┐
                  │  Analytics   │◀────────│   Redis    │
                  │   Service    │         │   Cache    │
                  │    (8083)    │         │   (6379)   │
                  └──────────────┘         └────────────┘

Each service has its own PostgreSQL database
```

### Service Ports

| Service | Port | Database |
|---------|------|----------|
| API Gateway | 8080 | - |
| User Service | 8081 | users_db (5433) |
| Workout Service | 8082 | workout_db (5434) |
| Analytics Service | 8083 | analytics_db (5435) |
| Exercise Service | 8084 | exercise_db (5436) |
| Redis | 6379 | - |

### Service Responsibilities

**API Gateway (8080)**
- Single entry point for all client requests
- JWT token validation
- Request routing to appropriate services
- Request ID generation and propagation
- Response aggregation
- Rate limiting (optional)

**User Service (8081)**
- User registration and authentication
- JWT token issuance and validation
- User profile management
- Role management (hardcoded as "user" for now)

**Exercise Service (8084)**
- Exercise catalog management
- Seed data from WGER API
- Exercise search and filtering
- Category, equipment, difficulty filtering
- Heavy caching (exercises rarely change)

**Workout Service (8082)**
- Workout CRUD operations
- Exercise logging with detailed sets
- Set types: warmup, normal, top, backoff, drop, failure, amrap, timed, distance
- Validates exercises via Exercise Service
- Triggers Analytics Service after workout creation

**Analytics Service (8083)**
- Real-time workout processing
- Performance history tracking
- Personal records detection and management
- Progress snapshots (weekly/monthly aggregates)
- Statistics generation and caching

**Redis (6379)**
- Session token caching
- Exercise catalog caching
- Recent workout data caching
- Analytics results caching

---

## Database Schemas

### User Service (users_db)

```sql
-- Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100),
    role VARCHAR(20) DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);

-- Seed admin user (optional)
INSERT INTO users (email, password_hash, name, role) 
VALUES ('admin@gym.com', '$2a$10$...', 'Admin User', 'admin');
```

**JWT Token Structure:**
```json
{
  "user_id": 1,
  "email": "john@example.com",
  "name": "John Doe",
  "role": "user",
  "exp": 1708171200,
  "iat": 1708084800,
  "iss": "user-service"
}
```

### Exercise Service (exercise_db)

```sql
-- Exercises catalog
CREATE TABLE exercises (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    instructions TEXT,
    category VARCHAR(50),                         -- 'strength', 'cardio', 'flexibility'
    primary_muscles JSON,                         -- ['chest', 'triceps']
    secondary_muscles JSON,                       -- ['shoulders']
    equipment VARCHAR(100),                       -- 'barbell', 'dumbbell', 'bodyweight', 'machine'
    difficulty VARCHAR(20),                       -- 'beginner', 'intermediate', 'advanced'
    rep_type VARCHAR(20) DEFAULT 'count',         -- 'count', 'time', 'distance'
    unit VARCHAR(20),                             -- 'reps', 'seconds', 'meters'
    external_source VARCHAR(50),                  -- 'wger', 'manual'
    external_id VARCHAR(100),                     -- ID from external API
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_exercises_category ON exercises(category);
CREATE INDEX idx_exercises_slug ON exercises(slug);
CREATE INDEX idx_exercises_equipment ON exercises(equipment);
CREATE INDEX idx_exercises_difficulty ON exercises(difficulty);
CREATE INDEX idx_exercises_muscles ON exercises USING GIN(primary_muscles);

-- Seed data: Manual exercises for Day 1-5
INSERT INTO exercises (name, slug, category, primary_muscles, equipment, difficulty, rep_type, unit, external_source) VALUES
('Barbell Bench Press', 'barbell-bench-press', 'strength', '["chest","triceps"]', 'barbell', 'intermediate', 'count', 'reps', 'manual'),
('Barbell Squat', 'barbell-squat', 'strength', '["quadriceps","glutes"]', 'barbell', 'intermediate', 'count', 'reps', 'manual'),
('Barbell Deadlift', 'barbell-deadlift', 'strength', '["back","glutes","hamstrings"]', 'barbell', 'advanced', 'count', 'reps', 'manual'),
('Pull-ups', 'pull-ups', 'strength', '["lats","biceps"]', 'bodyweight', 'intermediate', 'count', 'reps', 'manual'),
('Running', 'running', 'cardio', '["legs","cardiovascular"]', 'none', 'beginner', 'distance', 'meters', 'manual'),
('Plank', 'plank', 'strength', '["core"]', 'bodyweight', 'beginner', 'time', 'seconds', 'manual');
```

### Workout Service (workout_db)

```sql
-- Workout header
CREATE TABLE workouts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    name VARCHAR(255),
    date DATE NOT NULL,
    duration_minutes INTEGER,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_workouts_user_date ON workouts(user_id, date DESC);
CREATE INDEX idx_workouts_user ON workouts(user_id);
CREATE INDEX idx_workouts_date ON workouts(date DESC);

-- Exercises in workout
CREATE TABLE workout_exercises (
    id SERIAL PRIMARY KEY,
    workout_id INTEGER REFERENCES workouts(id) ON DELETE CASCADE,
    exercise_id INTEGER NOT NULL,                -- References Exercise Service
    order_index INTEGER,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_workout_exercises_workout ON workout_exercises(workout_id);
CREATE INDEX idx_workout_exercises_exercise ON workout_exercises(exercise_id);

-- Individual sets
CREATE TABLE exercise_sets (
    id SERIAL PRIMARY KEY,
    workout_exercise_id INTEGER REFERENCES workout_exercises(id) ON DELETE CASCADE,
    set_number INTEGER,
    reps DECIMAL(10,2),                          -- Supports time-based (12.5 seconds)
    weight DECIMAL(10,2),
    unit VARCHAR(10) DEFAULT 'kg',
    set_type VARCHAR(20) DEFAULT 'normal',       -- See valid types below
    notes TEXT,
    rpe INTEGER,                                 -- Rate of Perceived Exertion (1-10)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_exercise_sets_workout_exercise ON exercise_sets(workout_exercise_id);

-- Valid set_type values:
-- 'warmup'      - Warmup sets (excluded from analytics)
-- 'normal'      - Regular working sets
-- 'top'         - Top/peak sets (highest intensity)
-- 'backoff'     - Reduced weight after top sets
-- 'drop'        - Drop sets
-- 'failure'     - Sets to failure
-- 'amrap'       - As many reps as possible
-- 'approximation' - Proximity to failure (RPE-based)
-- 'timed'       - Time-based (planks, holds)
-- 'distance'    - Distance-based (running, rowing)
```

### Analytics Service (analytics_db)

```sql
-- Time-series performance history
CREATE TABLE exercise_performance_history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    exercise_id INTEGER NOT NULL,
    workout_id INTEGER NOT NULL,
    workout_date DATE NOT NULL,
    
    -- Aggregated metrics (working sets only, no warmups)
    max_weight DECIMAL(10,2),
    total_volume DECIMAL(10,2),                  -- Sum of (weight × reps)
    total_sets INTEGER,
    total_reps DECIMAL(10,2),
    avg_rpe DECIMAL(4,2),
    
    -- Raw data for reference
    raw_sets JSON,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_perf_history_user_exercise_date ON exercise_performance_history(user_id, exercise_id, workout_date DESC);
CREATE INDEX idx_perf_history_user_date ON exercise_performance_history(user_id, workout_date DESC);
CREATE INDEX idx_perf_history_exercise ON exercise_performance_history(exercise_id);

-- Recent data optimization
CREATE INDEX idx_perf_history_recent ON exercise_performance_history(user_id, workout_date DESC)
WHERE workout_date > CURRENT_DATE - INTERVAL '90 days';

-- Personal records
CREATE TABLE personal_records (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    exercise_id INTEGER NOT NULL,
    record_type VARCHAR(20),                     -- 'max_weight', 'max_reps', 'max_volume'
    value DECIMAL(10,2),
    unit VARCHAR(10),
    achieved_date DATE,
    workout_id INTEGER,
    previous_record DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(user_id, exercise_id, record_type)
);

CREATE INDEX idx_personal_records_user_exercise ON personal_records(user_id, exercise_id);

-- Weekly/monthly progress snapshots
CREATE TABLE progress_snapshots (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    period_type VARCHAR(20),                     -- 'week', 'month'
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    
    total_workouts INTEGER,
    total_volume DECIMAL(12,2),
    total_sets INTEGER,
    total_exercises INTEGER,
    avg_workout_duration INTEGER,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(user_id, period_type, period_start)
);

CREATE INDEX idx_progress_snapshots_user_period ON progress_snapshots(user_id, period_type, period_start DESC);

-- Current exercise statistics
CREATE TABLE exercise_stats (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    exercise_id INTEGER NOT NULL,
    
    current_max_weight DECIMAL(10,2),
    lifetime_volume DECIMAL(12,2),
    lifetime_sets INTEGER,
    lifetime_workouts INTEGER,
    last_performed_date DATE,
    trend_30d VARCHAR(20),                       -- 'improving', 'stable', 'declining'
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(user_id, exercise_id)
);

CREATE INDEX idx_exercise_stats_user ON exercise_stats(user_id);
CREATE INDEX idx_exercise_stats_exercise ON exercise_stats(exercise_id);
```

---

## API Response Structure

### Response Envelope (FINAL VERSION)

```go
type APIResponse struct {
    Data      []interface{} `json:"data"`              // ALWAYS an array
    Errors    []APIError    `json:"errors,omitempty"`  // Array for multi-field errors
    Meta      *Meta         `json:"meta,omitempty"`
    Timestamp int64         `json:"timestamp"`
}

type APIError struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Field   string                 `json:"field,omitempty"`    // For validation errors
    Details map[string]interface{} `json:"details,omitempty"`
}

type Meta struct {
    RequestID  string      `json:"request_id"`         // For distributed tracing
    Page       *Pagination `json:"page,omitempty"`
}

type Pagination struct {
    CurrentPage int `json:"current_page"`
    PageSize    int `json:"page_size"`
    TotalPages  int `json:"total_pages"`
    TotalItems  int `json:"total_items"`
}
```

### HTTP Status Codes

**Success (2xx)**
```
200 OK              - Successful GET, PUT, DELETE
201 Created         - Successful POST
202 Accepted        - Async processing started
204 No Content      - Successful DELETE with no body
```

**Client Errors (4xx)**
```
400 Bad Request           - Validation error
401 Unauthorized          - Missing/invalid auth
403 Forbidden             - Insufficient permissions
404 Not Found             - Resource doesn't exist
409 Conflict              - Duplicate resource
422 Unprocessable Entity  - Business rule violation
429 Too Many Requests     - Rate limit exceeded
```

**Server Errors (5xx)**
```
500 Internal Server Error - Unexpected error
502 Bad Gateway           - Invalid upstream response
503 Service Unavailable   - Service down
504 Gateway Timeout       - Upstream timeout
```

### Success Response Examples

**Single Resource**
```json
GET /api/users/1
Status: 200 OK

{
  "data": [
    {
      "id": 1,
      "email": "john@example.com",
      "name": "John Doe"
    }
  ],
  "meta": {
    "request_id": "req_abc123"
  },
  "timestamp": 1708084800
}
```

**Collection**
```json
GET /api/workouts?page=1&limit=10
Status: 200 OK

{
  "data": [
    {"id": 15, "name": "Push Day A", "date": "2025-02-16"},
    {"id": 14, "name": "Leg Day", "date": "2025-02-14"}
  ],
  "meta": {
    "request_id": "req_def456",
    "page": {
      "current_page": 1,
      "page_size": 10,
      "total_pages": 5,
      "total_items": 47
    }
  },
  "timestamp": 1708084800
}
```

**Empty Result**
```json
GET /api/workouts?user_id=999
Status: 200 OK

{
  "data": [],
  "meta": {
    "request_id": "req_ghi789",
    "page": {
      "current_page": 1,
      "page_size": 10,
      "total_pages": 0,
      "total_items": 0
    }
  },
  "timestamp": 1708084800
}
```

### Error Response Examples

**Single Validation Error**
```json
POST /api/workouts
Status: 400 Bad Request

{
  "data": [],
  "errors": [
    {
      "code": "VALIDATION_ERROR",
      "message": "Date is required",
      "field": "date",
      "details": {
        "constraint": "required",
        "received": null
      }
    }
  ],
  "meta": {
    "request_id": "req_jkl012"
  },
  "timestamp": 1708084800
}
```

**Multiple Validation Errors**
```json
POST /api/users/register
Status: 400 Bad Request

{
  "data": [],
  "errors": [
    {
      "code": "VALIDATION_ERROR",
      "message": "Invalid email format",
      "field": "email",
      "details": {"constraint": "email", "received": "invalid-email"}
    },
    {
      "code": "VALIDATION_ERROR",
      "message": "Password must be at least 8 characters",
      "field": "password",
      "details": {"constraint": "min_length", "required": 8, "received": 3}
    }
  ],
  "meta": {
    "request_id": "req_mno345"
  },
  "timestamp": 1708084800
}
```

**Nested Field Validation**
```json
POST /api/workouts
Status: 400 Bad Request

{
  "data": [],
  "errors": [
    {
      "code": "VALIDATION_ERROR",
      "message": "Reps must be positive",
      "field": "exercises[0].sets[0].reps",
      "details": {"constraint": "min", "min_value": 0, "received": -5}
    }
  ],
  "meta": {
    "request_id": "req_pqr678"
  },
  "timestamp": 1708084800
}
```

**Service Unavailable**
```json
GET /api/stats/progress
Status: 503 Service Unavailable

{
  "data": [],
  "errors": [
    {
      "code": "SERVICE_UNAVAILABLE",
      "message": "Analytics service is temporarily unavailable",
      "details": {
        "service": "analytics-service",
        "retry_after": 30
      }
    }
  ],
  "meta": {
    "request_id": "req_stu901"
  },
  "timestamp": 1708084800
}
```

### Error Codes (Standard)

```go
const (
    // Client Errors (4xx)
    ErrCodeValidation         = "VALIDATION_ERROR"
    ErrCodeUnauthorized       = "UNAUTHORIZED"
    ErrCodeForbidden          = "FORBIDDEN"
    ErrCodeNotFound           = "NOT_FOUND"
    ErrCodeConflict           = "CONFLICT"
    ErrCodeBusinessRule       = "BUSINESS_RULE_VIOLATION"
    ErrCodeRateLimit          = "RATE_LIMIT_EXCEEDED"
    ErrCodeInvalidParameter   = "INVALID_PARAMETER"
    
    // Server Errors (5xx)
    ErrCodeInternal           = "INTERNAL_ERROR"
    ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
    ErrCodeTimeout            = "TIMEOUT"
    ErrCodeDatabaseError      = "DATABASE_ERROR"
)
```

### Whitelisted Query Parameters

**Global (Pagination & Sorting)**
```
?page=1           - Page number (default: 1, min: 1)
?limit=20         - Items per page (default: 10, min: 1, max: 100)
?sort=field:order - Sort by field (e.g., date:desc,name:asc)
```

**Workout Endpoints**
```
GET /api/workouts

Allowed:
  ?user_id=1            - Filter by user
  ?date_from=2025-02-01 - Start date (YYYY-MM-DD)
  ?date_to=2025-02-16   - End date (YYYY-MM-DD)
  ?page=1
  ?limit=10
  ?sort=date:desc
```

**Exercise Endpoints**
```
GET /api/exercises

Allowed:
  ?category=strength         - Filter by category [strength, cardio, flexibility]
  ?equipment=barbell         - Filter by equipment
  ?difficulty=intermediate   - Filter by difficulty [beginner, intermediate, advanced]
  ?search=bench              - Text search (min 3 chars)
  ?muscles=chest             - Filter by muscle group
  ?page=1
  ?limit=10
  ?sort=name:asc
```

**Analytics Endpoints**
```
GET /api/stats/exercise/:id/history
Allowed:
  ?range=3months    - Time range [1month, 3months, 6months, 1year, all]
  ?unit=week        - Aggregation [day, week, month]

GET /api/stats/progress
Allowed:
  ?period=weekly    - Aggregation [weekly, monthly]
  ?weeks=12         - Number of periods (1-52)
```

---

## Service Specifications

### API Gateway

**Port:** 8080

**Responsibilities:**
- Entry point for all requests
- JWT validation
- Request routing
- Request ID generation
- Response standardization

**Routes:**
```
POST   /api/users/register        → User Service
POST   /api/users/login           → User Service
GET    /api/users/:id             → User Service

GET    /api/exercises             → Exercise Service
GET    /api/exercises/:id         → Exercise Service
GET    /api/exercises/search      → Exercise Service

POST   /api/workouts              → Workout Service
GET    /api/workouts              → Workout Service
GET    /api/workouts/:id          → Workout Service
PUT    /api/workouts/:id          → Workout Service
DELETE /api/workouts/:id          → Workout Service

GET    /api/stats/exercise/:id/history  → Analytics Service
GET    /api/stats/personal-records      → Analytics Service
GET    /api/stats/progress               → Analytics Service
```

### User Service

**Port:** 8081
**Database:** users_db (PostgreSQL on 5433)

**Endpoints:**
```
POST   /api/users/register
Body: { "email": "...", "password": "...", "name": "..." }
Response: { "data": [{ "id": 1, "email": "...", "name": "..." }] }

POST   /api/users/login
Body: { "email": "...", "password": "..." }
Response: { "data": [{ "token": "jwt_token", "user": { ... } }] }

GET    /api/users/:id
Headers: Authorization: Bearer <token>
Response: { "data": [{ "id": 1, "email": "...", "name": "..." }] }
```

### Exercise Service

**Port:** 8084
**Database:** exercise_db (PostgreSQL on 5436)

**Endpoints:**
```
GET    /api/exercises
Query: ?category=strength&difficulty=intermediate&page=1&limit=10
Response: { "data": [{ "id": 1, "name": "Bench Press", ... }], "meta": { "page": { ... } } }

GET    /api/exercises/:id
Response: { "data": [{ "id": 1, "name": "Bench Press", "description": "...", ... }] }

GET    /api/exercises/search
Query: ?q=bench
Response: { "data": [{ "id": 1, "name": "Bench Press", ... }] }
```

**Redis Caching:**
```
Key: exercise:id:{id}              TTL: 24 hours
Key: exercise:list:category:{cat}  TTL: 1 hour
Key: exercise:search:{query}       TTL: 30 minutes
```

### Workout Service

**Port:** 8082
**Database:** workout_db (PostgreSQL on 5434)

**Endpoints:**
```
POST   /api/workouts
Headers: Authorization: Bearer <token>
Body: {
  "name": "Push Day A",
  "date": "2025-02-16",
  "duration_minutes": 75,
  "notes": "Great session",
  "exercises": [
    {
      "exercise_id": 1,
      "order_index": 1,
      "notes": "Paused reps",
      "sets": [
        {
          "set_number": 1,
          "reps": 8,
          "weight": 60,
          "unit": "kg",
          "set_type": "warmup",
          "rpe": 4
        },
        {
          "set_number": 2,
          "reps": 5,
          "weight": 100,
          "unit": "kg",
          "set_type": "top",
          "rpe": 9,
          "notes": "New PR!"
        }
      ]
    }
  ]
}
Response: { "data": [{ "id": 123, "name": "Push Day A", ... }] }

GET    /api/workouts
Query: ?date_from=2025-02-01&date_to=2025-02-16&page=1&limit=10
Response: { "data": [{ "id": 123, ... }], "meta": { "page": { ... } } }

GET    /api/workouts/:id
Response: { "data": [{ "id": 123, "exercises": [{ "sets": [...] }] }] }

PUT    /api/workouts/:id
Body: { "name": "Updated Name", ... }
Response: { "data": [{ "id": 123, ... }] }

DELETE /api/workouts/:id
Response: { "data": [{ "id": 123, "deleted": true }] }
```

**After Workout Creation:**
```go
// Trigger Analytics Service (async)
go func() {
    http.Post("http://analytics-service:8083/api/analytics/process-workout",
        "application/json",
        bytes.NewBuffer([]byte(`{"workout_id": 123}`)))
}()
```

### Analytics Service

**Port:** 8083
**Database:** analytics_db (PostgreSQL on 5435)

**Endpoints:**
```
POST   /internal/analytics/process-workout
Body: { "workout_id": 123 }
Response: { "data": [{ "processed": true }] }
(Called by Workout Service, not exposed via Gateway)

GET    /api/stats/exercise/:id/history
Query: ?range=3months&unit=week
Response: {
  "data": [{
    "exercise_id": 1,
    "data_points": [
      { "date": "2025-02-16", "max_weight": 100, "total_volume": 2450 },
      ...
    ]
  }]
}

GET    /api/stats/personal-records
Response: {
  "data": [{
    "records": [
      {
        "exercise_id": 1,
        "exercise_name": "Bench Press",
        "max_weight": { "value": 110, "unit": "kg", "achieved_date": "2025-02-10" }
      }
    ]
  }]
}

GET    /api/stats/progress
Query: ?period=weekly&weeks=12
Response: {
  "data": [{
    "weeks": [
      { "week_start": "2025-02-10", "total_workouts": 4, "total_volume": 8500 }
    ]
  }]
}
```

**Redis Caching:**
```
Key: analytics:user:{id}:history:exercise:{id}:90d    TTL: 1 hour
Key: analytics:user:{id}:personal_records             TTL: 24 hours
Key: analytics:user:{id}:progress:weekly              TTL: 1 hour
```

---

## Analytics Data Flow

### Real-time Processing Flow

```
1. User completes workout
   ↓
2. Workout Service saves to database
   ↓
3. Workout Service triggers Analytics (async HTTP POST)
   ↓
4. Analytics Service receives workout_id
   ↓
5. Analytics fetches full workout from Workout Service API
   ↓
6. For each exercise in workout:
   ├─ Calculate metrics (excluding warmup sets)
   ├─ Store in exercise_performance_history
   ├─ Check for personal records
   ├─ Update personal_records if new PR
   └─ Update exercise_stats
   ↓
7. Update progress_snapshots (weekly aggregates)
   ↓
8. Cache results in Redis
   ↓
9. Analytics ready for queries
```

### Metrics Calculation Logic

```go
func calculateExerciseMetrics(exercise WorkoutExercise) ExerciseMetrics {
    var metrics ExerciseMetrics
    var workingSets []ExerciseSet
    
    // Filter out warmup sets
    for _, set := range exercise.Sets {
        if set.SetType == "warmup" {
            continue
        }
        workingSets = append(workingSets, set)
        
        // Track max weight
        if set.Weight > metrics.MaxWeight {
            metrics.MaxWeight = set.Weight
        }
        
        // Calculate volume (weight × reps)
        metrics.TotalVolume += set.Weight * set.Reps
        
        // Accumulate reps
        metrics.TotalReps += set.Reps
        
        // Track RPE
        if set.RPE > 0 {
            totalRPE += float64(set.RPE)
            rpeCount++
        }
    }
    
    metrics.TotalSets = len(workingSets)
    
    if rpeCount > 0 {
        metrics.AvgRPE = totalRPE / float64(rpeCount)
    }
    
    return metrics
}
```

---

## Day-by-Day Implementation Plan

### **DAY 1: Foundation + User Service** (Feb 16)

**Goal:** Project structure + working User Service

**Tasks:**
1. Create project structure
```
gym-tracker/
├── docker-compose.yml
├── .env
├── api-gateway/
│   ├── main.go
│   ├── go.mod
│   └── handlers/
├── user-service/
│   ├── main.go
│   ├── go.mod
│   ├── handlers/
│   ├── models/
│   └── repository/
├── exercise-service/
├── workout-service/
├── analytics-service/
└── shared/
    ├── response.go
    └── middleware.go
```

2. Create docker-compose.yml (see Docker Setup section)

3. Build User Service:
   x- Database connection
   x- User registration (bcrypt password hashing)
   x- User login (JWT generation)
   x- Get user profile endpoint
   x- Standardized API responses

**Deliverable:** Can register user and receive JWT token

**Testing:**
```bash
# Register
curl -X POST http://localhost:8081/api/users/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@gym.com","password":"password123","name":"Test User"}'

# Login
curl -X POST http://localhost:8081/api/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@gym.com","password":"password123"}'
```

---

### **DAY 2: API Gateway + Exercise Service** (Feb 17)

**Goal:** Gateway routing + Exercise catalog

**Tasks:**
1. Build API Gateway:
   - Request ID generation middleware
   x- JWT validation middleware
   x - Reverse proxy to User Service
   x - Reverse proxy to Exercise Service
   - Error handling and response standardization

2. Build Exercise Service:
   - Database connection
   - Seed 20-30 exercises manually
   - List exercises endpoint (with pagination)
   - Get exercise by ID
   - Search exercises
   - Redis caching implementation

3. Connect services through Gateway

**Deliverable:** Can login and browse exercises via Gateway

**Testing:**
```bash
# Login via Gateway
curl -X POST http://localhost:8080/api/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@gym.com","password":"password123"}'

# Get exercises
curl http://localhost:8080/api/exercises?category=strength&page=1&limit=10

# Search exercises
curl "http://localhost:8080/api/exercises/search?q=bench"
```

---

### **DAY 3: Workout Service** (Feb 18)

**Goal:** Create and view workouts with sets

**Tasks:**
1. Build Workout Service database (3 tables)
2. Implement workout creation:
   - Validate user_id from JWT
   - Validate exercise_id with Exercise Service
   - Save nested exercises and sets
   - Support all set_types
   - Trigger Analytics Service (async)

3. Implement workout retrieval:
   - Get workout by ID (with full details)
   - List user workouts (paginated)
   - Filter by date range

4. Connect to API Gateway

**Deliverable:** Can create and view workouts with detailed sets

**Testing:**
```bash
# Create workout (see full body in Service Specifications)
curl -X POST http://localhost:8080/api/workouts \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d @workout.json

# List workouts
curl "http://localhost:8080/api/workouts?date_from=2025-02-01" \
  -H "Authorization: Bearer <token>"

# Get specific workout
curl http://localhost:8080/api/workouts/123 \
  -H "Authorization: Bearer <token>"
```

---

### **DAY 4: Analytics Service** (Feb 19)

**Goal:** Real-time analytics processing

**Tasks:**
1. Build Analytics Service:
   - Receive workout completion events
   - Fetch workout details from Workout Service
   - Calculate metrics (filter warmups)
   - Store in exercise_performance_history
   - Check and update personal records
   - Update exercise_stats

2. Implement analytics endpoints:
   - Get exercise history
   - Get personal records
   - Get progress snapshots

3. Redis caching for analytics

**Deliverable:** Analytics generated after workout creation

**Testing:**
```bash
# Create workout (triggers analytics)
# Wait 2 seconds

# Check personal records
curl http://localhost:8080/api/stats/personal-records \
  -H "Authorization: Bearer <token>"

# Check exercise history
curl "http://localhost:8080/api/stats/exercise/1/history?range=3months" \
  -H "Authorization: Bearer <token>"
```

---

### **DAY 5: Integration & Caching** (Feb 20)

**Goal:** Polish service communication + optimize caching

**Tasks:**
1. Enhance Redis caching:
   - Cache exercise catalog aggressively
   - Cache recent workouts
   - Cache analytics results
   - Implement cache invalidation

2. Add progress snapshots:
   - Weekly aggregates
   - Monthly summaries

3. Service resilience:
   - Retry logic for HTTP calls
   - Timeout handling
   - Graceful degradation

4. Request ID tracing across all services

**Deliverable:** Fast responses with proper caching

---

### **DAY 6: WGER Integration + Polish** (Feb 21)

**Goal:** Import exercises from WGER API

**Tasks:**
1. WGER API integration:
   - Fetch exercises from https://wger.de/api/v2/exercise/
   - Transform data to our schema
   - Seed database with 100+ exercises
   - Handle pagination

2. Improve Exercise Service:
   - Advanced filtering
   - Better search
   - Category/equipment/difficulty filters

3. Error handling improvements:
   - Consistent error codes
   - Better error messages
   - Structured logging

4. Health check endpoints for all services

**Deliverable:** Rich exercise catalog from WGER

---

### **DAY 7: Docker + Testing + Documentation** (Feb 22)

**Goal:** Production-ready containerization

**Tasks:**
1. Create Dockerfiles for all services
2. Update docker-compose.yml for full stack
3. Environment variable management
4. Write integration tests
5. Create API documentation
6. Write setup instructions
7. Document lessons learned

**Deliverable:** Run entire system with `docker-compose up`

---

## Docker Setup

### docker-compose.yml

```yaml
version: '3.8'

services:
  # PostgreSQL databases
  users-db:
    image: postgres:15-alpine
    container_name: users-db
    environment:
      POSTGRES_DB: users_db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5433:5432"
    volumes:
      - users-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  workout-db:
    image: postgres:15-alpine
    container_name: workout-db
    environment:
      POSTGRES_DB: workout_db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5434:5432"
    volumes:
      - workout-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  analytics-db:
    image: postgres:15-alpine
    container_name: analytics-db
    environment:
      POSTGRES_DB: analytics_db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5435:5432"
    volumes:
      - analytics-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  exercise-db:
    image: postgres:15-alpine
    container_name: exercise-db
    environment:
      POSTGRES_DB: exercise_db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5436:5432"
    volumes:
      - exercise-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Redis
  redis:
    image: redis:7-alpine
    container_name: redis
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Services (Day 7 - uncomment when Dockerfiles ready)
  # user-service:
  #   build: ./user-service
  #   container_name: user-service
  #   ports:
  #     - "8081:8081"
  #   environment:
  #     - DB_HOST=users-db
  #     - DB_PORT=5432
  #     - DB_USER=postgres
  #     - DB_PASSWORD=postgres
  #     - DB_NAME=users_db
  #     - JWT_SECRET=${JWT_SECRET}
  #   depends_on:
  #     users-db:
  #       condition: service_healthy
  #   restart: unless-stopped

volumes:
  users-data:
  workout-data:
  analytics-data:
  exercise-data:
  redis-data:
```

### Example Dockerfile (User Service)

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .

EXPOSE 8081

CMD ["./main"]
```

---

## Environment Configuration

### .env file

```env
# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-in-production

# Database Configuration (Development)
USERS_DB_HOST=localhost
USERS_DB_PORT=5433
USERS_DB_USER=postgres
USERS_DB_PASSWORD=postgres
USERS_DB_NAME=users_db

WORKOUT_DB_HOST=localhost
WORKOUT_DB_PORT=5434
WORKOUT_DB_USER=postgres
WORKOUT_DB_PASSWORD=postgres
WORKOUT_DB_NAME=workout_db

ANALYTICS_DB_HOST=localhost
ANALYTICS_DB_PORT=5435
ANALYTICS_DB_USER=postgres
ANALYTICS_DB_PASSWORD=postgres
ANALYTICS_DB_NAME=analytics_db

EXERCISE_DB_HOST=localhost
EXERCISE_DB_PORT=5436
EXERCISE_DB_USER=postgres
EXERCISE_DB_PASSWORD=postgres
EXERCISE_DB_NAME=exercise_db

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379

# Service URLs (Development)
USER_SERVICE_URL=http://localhost:8081
EXERCISE_SERVICE_URL=http://localhost:8084
WORKOUT_SERVICE_URL=http://localhost:8082
ANALYTICS_SERVICE_URL=http://localhost:8083

# WGER API (Day 6)
WGER_API_URL=https://wger.de/api/v2
```

---

## Testing Strategy

### Unit Tests

```go
// Example: Test metrics calculation
func TestCalculateVolume(t *testing.T) {
    sets := []ExerciseSet{
        {Reps: 10, Weight: 100, SetType: "warmup"},
        {Reps: 8, Weight: 120, SetType: "normal"},
        {Reps: 6, Weight: 130, SetType: "top"},
    }
    
    volume := calculateVolume(sets)
    expected := 8*120 + 6*130 // Excluding warmup
    
    if volume != expected {
        t.Errorf("Expected volume %f, got %f", expected, volume)
    }
}
```

### Integration Tests

```bash
#!/bin/bash
# test-integration.sh

# 1. Register user
TOKEN=$(curl -s -X POST http://localhost:8080/api/users/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@gym.com","password":"password123","name":"Test"}' \
  | jq -r '.data[0].token')

echo "Token: $TOKEN"

# 2. List exercises
curl -s http://localhost:8080/api/exercises | jq '.data | length'

# 3. Create workout
WORKOUT_ID=$(curl -s -X POST http://localhost:8080/api/workouts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d @test-workout.json \
  | jq -r '.data[0].id')

echo "Workout ID: $WORKOUT_ID"

# 4. Wait for analytics
sleep 2

# 5. Check stats
curl -s http://localhost:8080/api/stats/personal-records \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.data[0].records'
```

---

## Quick Start Commands

### Development (Day 1-6)

```bash
# Start infrastructure
docker-compose up -d users-db workout-db analytics-db exercise-db redis

# Run services locally
cd user-service && go run main.go &
cd exercise-service && go run main.go &
cd workout-service && go run main.go &
cd analytics-service && go run main.go &
cd api-gateway && go run main.go &

# Check all services
curl http://localhost:8081/health
curl http://localhost:8082/health
curl http://localhost:8083/health
curl http://localhost:8084/health
curl http://localhost:8080/health
```

### Production (Day 7)

```bash
# Build and start all services
docker-compose up --build

# Check logs
docker-compose logs -f user-service

# Stop all
docker-compose down

# Clean restart
docker-compose down -v
docker-compose up --build
```

---

## Go Dependencies

```bash
# Each service needs:
go get github.com/gorilla/mux              # HTTP routing
go get github.com/lib/pq                   # PostgreSQL driver
go get github.com/go-redis/redis/v8        # Redis client
go get github.com/golang-jwt/jwt/v5        # JWT handling
go get github.com/joho/godotenv            # Environment variables
go get golang.org/x/crypto/bcrypt          # Password hashing
go get github.com/google/uuid              # UUID generation
```

---

## Success Criteria

### Minimum Viable Product (Must Complete)
- ✅ 5 services running independently
- ✅ User registration and authentication with JWT
- ✅ Browse exercise catalog (manual or WGER)
- ✅ Create workouts with detailed sets
- ✅ View workout history
- ✅ View personal records
- ✅ View exercise performance history
- ✅ Redis caching active
- ✅ All services containerized

### Bonus Features (If Time Permits)
- ⭐ WGER API integration complete
- ⭐ Circuit breaker pattern
- ⭐ Request tracing visualization
- ⭐ Background job for weekly summaries
- ⭐ Comprehensive test coverage
- ⭐ API documentation (Swagger)

---

## Key Learning Outcomes

By completing this project, you will understand:

✅ **Service Decomposition** - How to split functionality into services
✅ **Database Per Service** - Data isolation and ownership
✅ **Service Communication** - HTTP REST between services
✅ **API Design** - Consistent response structures
✅ **Distributed Tracing** - Request IDs across services
✅ **Caching Strategies** - Redis for performance
✅ **Real-time Processing** - Async event handling
✅ **Error Handling** - Graceful degradation
✅ **Containerization** - Docker and Docker Compose
✅ **Query Validation** - Security best practices

---

## Final Checklist

**Before Starting:**
- [ ] Install Go 1.21+
- [ ] Install Docker and Docker Compose
- [ ] Install PostgreSQL client (psql)
- [ ] Install Redis client (redis-cli)
- [ ] Install curl or Postman

**Day 1:**
- [ ] Project structure created
- [ ] docker-compose.yml working
- [ ] User Service: register, login, JWT working

**Day 2:**
- [ ] API Gateway routing to User Service
- [ ] Exercise Service with seed data
- [ ] Redis caching implemented

**Day 3:**
- [ ] Workout Service CRUD complete
- [ ] Nested exercises and sets working
- [ ] Analytics trigger implemented

**Day 4:**
- [ ] Analytics processing workouts
- [ ] Personal records detection
- [ ] Performance history tracking

**Day 5:**
- [ ] All caching optimized
- [ ] Progress snapshots working
- [ ] Request ID tracing complete

**Day 6:**
- [ ] WGER integration (optional)
- [ ] Error handling polished
- [ ] Health checks added

**Day 7:**
- [ ] All Dockerfiles created
- [ ] docker-compose up works
- [ ] Integration tests passing
- [ ] Documentation complete

---

## Resources

**External APIs:**
- WGER: https://wger.de/api/v2/

**Documentation:**
- Go HTTP: https://gobyexample.com/http-servers
- PostgreSQL: https://github.com/lib/pq
- Redis: https://redis.uptrace.dev/guide/
- JWT: https://pkg.go.dev/github.com/golang-jwt/jwt/v5

**Patterns:**
- Microservices Patterns (Martin Fowler)
- The Twelve-Factor App
- API Design Best Practices

---

**Good luck with your implementation! 🚀**

Remember: Focus on learning concepts over perfect code. It's okay if you don't finish everything - the journey is the goal!
