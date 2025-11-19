FROM python:3.12-slim

WORKDIR /app

# Install runtime dependencies
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY . .

# Ensure start script is executable
RUN chmod +x start.sh

# Expose default port
EXPOSE 8080

# Environment defaults (can be overridden at runtime)
ENV APP_HOST=0.0.0.0 \
    APP_PORT=8080 \
    APP_WORKERS=4

# Use Gunicorn to run the Flask app
CMD ["./start.sh"]
