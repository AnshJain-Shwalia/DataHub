import express from 'express';
import { EventEmitter } from 'events';
import { Server } from 'http';

export class OAuthServer extends EventEmitter {
  private app: express.Application;
  private server: Server | null = null;
  private readonly port = 9753;

  constructor() {
    super();
    this.app = express();
    this.setupRoutes();
  }

  private setupRoutes(): void {
    // Health check
    this.app.get('/health', (req, res) => {
      res.json({ status: 'ok', port: this.port });
    });

    // OAuth callback handlers
    this.app.get('/auth/:provider/callback', (req, res) => {
      const { provider } = req.params;
      const { code, state, error, error_description } = req.query;

      // Send a response that closes the browser window
      res.send(`
        <!DOCTYPE html>
        <html>
        <head>
          <title>Authentication Complete</title>
          <style>
            body { 
              font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
              display: flex;
              justify-content: center;
              align-items: center;
              height: 100vh;
              margin: 0;
              background-color: #f7fafc;
            }
            .message {
              text-align: center;
              padding: 2rem;
              background: white;
              border-radius: 8px;
              box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
            }
            .success { color: #38a169; }
            .error { color: #e53e3e; }
          </style>
        </head>
        <body>
          <div class="message">
            ${error ? 
              `<h2 class="error">Authentication Failed</h2><p>${error_description || error}</p>` :
              `<h2 class="success">Authentication Successful</h2><p>You can close this window.</p>`
            }
          </div>
          <script>
            // Auto-close after 2 seconds
            setTimeout(() => {
              window.close();
            }, 2000);
          </script>
        </body>
        </html>
      `);

      // Emit events to main process
      if (error) {
        this.emit('auth-error', provider, error_description || error);
      } else if (code && state) {
        this.emit('auth-success', provider, code, state);
      } else {
        this.emit('auth-error', provider, 'Missing code or state parameter');
      }
    });

    // Catch-all for unhandled routes
    this.app.get('*', (req, res) => {
      res.status(404).json({ 
        error: 'Not found',
        message: 'This is the DataHub OAuth callback server. Only OAuth callbacks are handled here.'
      });
    });
  }

  public async start(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.server) {
        resolve(); // Already running
        return;
      }

      this.server = this.app.listen(this.port, 'localhost', (err?: Error) => {
        if (err) {
          reject(new Error(`Failed to start OAuth server on port ${this.port}: ${err.message}`));
        } else {
          console.log(`OAuth callback server started on http://localhost:${this.port}`);
          resolve();
        }
      });

      this.server.on('error', (err: any) => {
        if (err.code === 'EADDRINUSE') {
          reject(new Error(`Port ${this.port} is already in use. Please close any applications using this port and try again.`));
        } else {
          reject(new Error(`OAuth server error: ${err.message}`));
        }
      });
    });
  }

  public async stop(): Promise<void> {
    return new Promise((resolve) => {
      if (this.server) {
        this.server.close(() => {
          console.log('OAuth callback server stopped');
          this.server = null;
          resolve();
        });
      } else {
        resolve();
      }
    });
  }

  public isRunning(): boolean {
    return this.server !== null;
  }
}