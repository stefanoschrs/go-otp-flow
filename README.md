# Go OTP Flow

### ENV
- OTPF_ISSUER (optional)
- PORT (optional)

### Generate
- GUI 
    ```
    0.0.0.0:8080/generate
    ```
- Headless
    ```
    curl "0.0.0.0:8080/generate?id=123&headless=true"
    curl "0.0.0.0:8080/generate?id=123&headless=true&issuer=google.com"
    curl "0.0.0.0:8080/generate?id=123&headless=true&type=image" | base64 -d > /tmp/img.png
    ```

### Validate
```
curl -X POST -H "content-type: application/json" -d '{"id":"123","token":"123"}' "0.0.0.0:8080/validate"
curl -X POST -H "content-type: application/json" -d '{"id":"123","token":"123","issuer":"google.com"}' "0.0.0.0:8080/validate"
```
