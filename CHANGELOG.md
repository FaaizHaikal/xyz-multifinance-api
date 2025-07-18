# Release Notes – PT XYZ Multifinance Core API v1.0.1  
**Date:** July 19, 2025

---

## Overview

This patch addresses critical bugs and improves stability after the initial `v1.0.0` release.  
**Note:** `v1.0.0` is now deprecated and will no longer be maintained.

---

## 🐛 Bug Fixes & Stability Improvements

- **Customer Login Fix**: Password hash was not loaded correctly due to JSON and GORM behavior. Fixed.
- **DB Connection Error**: Fixed "sql: database is closed" by adjusting GORM connection settings.
- **Transaction Data Handling**: Ensured consistent UUID generation and valid test data preloading.
- **Redis Initialization**: Fixed Redis client setup and removed nil pointer panics in rate limiting.
- **Table Naming**: Fixed migration table names to match GORM's pluralization.

---

## ⚙️ Technical Updates

- **GORM Logging**: SQL queries now logged properly for easier debugging.
- **Isolated Tests**: Unit tests for transactions now use separate SQLite instances to avoid contamination.

---

## 📝 What’s Next

Development will continue on new features as outlined in the `v1.0.0` release notes.

---
---

# Release Notes – PT XYZ Multifinance Core API v1.0.0
**Date:** July 18, 2025

---

## Overview

This release introduced the first version of the PT XYZ Multifinance Core API. **Please note: This version has been superseded by `v1.0.1` and is no longer recommended for use due to critical bug fixes available in the newer version.**

---

## 🚀 New Features

### Customer Management
- Register and manage customer profiles with NIK, personal details, and identity photos.
- Fetch customers by ID or NIK.

### Credit Limit Management
- Add and manage credit limits across multiple tenors (1, 2, 3, 6 months).
- Retrieve all credit limits for a given customer.

### Transaction Handling
- Create new transactions
- Automatically deduct from the appropriate credit limit.
- Concurrency-safe and ACID-compliant to prevent race conditions.
- Fetch transaction details by contract number or customer ID.

---

## 🔐 Security

- JWT authentication with access and refresh tokens
- Redis-based rate limiting
- SQL injection protection via prepared statements and input validation
- Secure environment-based configuration

---

## ⚙️ Technical Highlights

- Developed in Go using Clean Architecture
- Uses GORM for ORM and Redis for caching/rate limiting
- Fully covered by unit tests for all business logic

---

## 📝 What’s Next

Upcoming releases will expand into more financial products, integrate with external systems, and introduce reporting features.

---

## 📎 Documentation

- For the **architecture diagram** and **entity relationship diagram (ERD)**, please refer to the `/docs` directory.
- Diagrams are also embedded in the [README.md](./README.md) for quick reference.