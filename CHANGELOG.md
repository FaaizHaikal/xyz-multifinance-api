# Release Notes – PT XYZ Multifinance Core API v1.0.0  
**Date:** July 18, 2025

---

## Overview

This release introduces the first version of the PT XYZ Multifinance Core API. It replaces parts of the legacy monolith with a new modular backend built in Go, focusing on customer, credit limit, and transaction management.

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

## 📎 Documentation

- For the **architecture diagram** and **entity relationship diagram (ERD)**, please refer to the `/docs` directory.
- Diagrams are also embedded in the [README.md](./README.md) for quick reference.
