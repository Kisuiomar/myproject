# User Management System

This project is a simple user management system that allows you to:
- Add new users
- View a list of users
- Delete users
- Update user information
- Search users by ID, Name, or Email

## Requirements

Before running the project, you need to have the following tools installed:

- Go (Golang) - version 1.18 or higher
- PostgreSQL - used as the database
- SMTP Server credentials (for sending email notifications)

## Setup

### 1. Install Dependencies

To install the required Go packages for this project, run the following commands:

```bash
go mod init user-management
go get gorm.io/gorm
go get gorm.io/driver/postgres
