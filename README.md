# Email Campaign Management System

## Overview
The Email Campaign Management System is a web-based application designed to facilitate the composition, scheduling, and sending of marketing emails. It includes features such as email templates, campaign management, user subscriptions, and a streamlined interface for efficient email communication.

## Features
- **Email Composition**: Compose and format emails using the Quill text editor.
- **Campaign Management**: Create, manage, and delete campaigns with subscriber lists.
- **Email Templates**: Save and reuse email templates for consistent messaging.
- **Recipient Management**: Add individual recipients or select predefined mailing lists.
- **Attachment Support**: Attach files to emails before sending.
- **Scheduling**: Send emails immediately or schedule them for later.
- **Notifications**: Get feedback on email sending status.
- **Database Management**: Store email templates, campaigns, and subscribers using SQLite.

## Installation & Setup
### Prerequisites
- Go 1.18+
- Node.js (optional for frontend development)
- SQLite

### Installation Steps
1. **Clone the repository:**
   ```sh
   git clone https://github.com/your-repo/email-campaign.git
   cd email-campaign
   ```
2. **Set up environment variables:**
   Create a `.env` file and configure SMTP credentials:
   ```sh
   SMTP_HOST=smtp.example.com
   SMTP_USER=your-email@example.com
   SMTP_PASSWORD=your-password
   ```
3. **Initialize the database:**
   ```sh
   go run main.go
   ```
   This will create necessary database tables automatically.
4. **Build and run the project:**
   ```sh
   ./build.sh
   go run .
   ```

## Usage
### Running the Application
- Open `http://localhost:8080` in a browser.
- Use the interface to compose and send emails.
- Create campaigns and add subscribers.
- Select email templates and manage attachments.

## File Structure
```
.
├── templates/        # HTML templates for email composition
├── JavaScript/       # Frontend scripts (campaign selection, email handling)
├── main.go           # Entry point for the Go server
├── DB.go             # Database operations
├── mail.go           # Email sending logic
├── handlers.go       # HTTP route handlers
├── build.sh          # Build script
├── tasks.json        # VS Code build tasks
├── .env              # Environment variables for SMTP
├── templates.db      # SQLite database file
```

## API Endpoints
### Templates
- `GET /templates` - Retrieve all email templates
- `POST /templates` - Save a new email template

### Campaigns
- `POST /campaigns` - Create a new campaign
- `GET /campaigns/list` - Retrieve all campaigns

### Email Composition
- `POST /composer` - Send an email

## Dependencies
- [Go](https://golang.org/)
- [SQLite](https://sqlite.org/)
- [HTMX](https://htmx.org/)
- [Quill](https://quilljs.com/)
- [Tailwind CSS](https://tailwindcss.com/)
- [Golang SQLite Driver](https://github.com/mattn/go-sqlite3)
