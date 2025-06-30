---
trigger: always_on
---

## 📦 Project Overview: GitHub-as-Cloud Storage

This is a desktop application (Electron) that treats **GitHub repositories as cloud storage** by pushing file chunks into user-owned GitHub repos. The backend, built with **Go using Gin, GORM and gin-swagger**, handles chunk metadata, repository management, and integrates with object storage (e.g. S3-compatible) as a temporary buffer.

---

**⚠️ Note: Some details are subject to change as development evolves.**

---

## 🧱 Architecture

- **Frontend**: Electron app.
- **Backend**: Go server using:
  - **Gin** (HTTP routing)
  - **GORM** (ORM for PostgreSQL)
  - **godotenv** (environment config for local/dev)
  - **gin-swagger** (auto-generated OpenAPI docs)
- **Authentication**:  
  - Google OAuth for main login.  
  - GitHub OAuth to link one or more GitHub accounts per user.
- **Storage Flow**:  
  - Temporary: S3-compatible object store (e.g. AWS S3, Cloudflare R2).  
  - Permanent: GitHub repositories.

---

## 🔼 Upload Process

1. User selects files/folders in the Electron app.
2. Files > 5MB are chunked.
3. Client uploads chunks to object storage via pre-signed URLs.
4. Metadata (chunk info, path, repo mapping) is stored in PostgreSQL via **GORM**.
5. Upload triggers a serverless function (Go or TypeScript):
   - Pushes chunk to GitHub.
   - Deletes chunk from storage.
   - Notifies backend.
6. Client polls backend (via **Gin** API) every 3 seconds for upload status.

---

## 🔽 Download Process

1. User selects files to retrieve.
2. Backend returns chunk-repo mappings.
3. Client performs Git sparse checkout in parallel.
4. Chunks are reassembled locally.

---

## 📚 Metadata Handling

- Stored in **PostgreSQL** using **GORM**: filenames, paths, chunk list, repo IDs, etc.
- **Redis** may be used for faster polling and lookup.
- Chunk order enforced for accurate reassembly.

---

## ⚙️ System Limits

- **Chunk size**: Max 5MB per chunk.
- **Per repo limit**: 500MB.
- **User quota**: Up to 500GB across ~1,000 GitHub repos.

---

## 🧰 Tech Stack

| Component       | Technology                          |
|----------------|-------------------------------------|
| Desktop App    | Electron                            |
| Backend        | Go (Golang) + Gin + GORM            |
| Env Config     | godotenv                            |
| API Docs       | gin-swagger                         |
| Object Storage | S3-compatible (e.g. R2, S3)          |
| Database       | PostgreSQL (+ optional Redis)       |
| Serverless     | Platform-agnostic (Go/TS)           |
| GitHub Access  | GitHub OAuth API                    |

---

## 📁 Folder Structure

Below is an outline of the folder structure for this project. This section serves as a reference for organizing code, assets, configurations, and documentation across both the frontend and backend.

Root Directory  
│  
├── client/               # Electron frontend  
├── backend/              # Go backend (Gin + GORM + Swagger)  
├── lambda/               # Serverless functions (Go/TS)  
└── README.md             # Project overview