# Frontend MVP Check & Postman Flow

Tài liệu này dùng để:

1. Liệt kê các chức năng thực tế của frontend.
2. Chỉ ra các API mà frontend đang gọi.
3. Ghi lại flow Postman cần test để kiểm tra xem MVP đã đủ luồng chính của business logic hay chưa.

## 1. Frontend đang có những chức năng gì

### 1.1. Màn hình đặt xe

Frontend chỉ có một form đặt xe chính với các chức năng sau:

- Nhập điểm đón bằng `latitude` và `longitude`.
- Nhập điểm trả bằng `latitude` và `longitude`.
- Chọn loại xe `economy` hoặc `premium`.
- Hiển thị loading khi đang gọi API.
- Hiển thị lỗi nếu tạo chuyến thất bại.
- Tự reset form về tọa độ mặc định sau khi đặt xe thành công.

Giá trị mặc định đang được set sẵn trong UI:

- Pickup: `40.7128, -74.0060`
- Dropoff: `40.7580, -73.9855`

### 1.2. Màn hình theo dõi chuyến đi

Sau khi tạo chuyến thành công, frontend chuyển sang màn hình tracking với các chức năng sau:

- Hiển thị thông tin tài xế.
- Hiển thị trạng thái chuyến đi.
- Hiển thị tọa độ pickup, current, dropoff.
- Hiển thị ETA còn lại nếu chuyến chưa hoàn tất.
- Hiển thị giá cước.
- Poll API mỗi 3 giây để cập nhật dữ liệu chuyến.
- Khi trip `completed`, hiện form rating + feedback.
- Có nút `Book Another Ride` để quay về luồng đặt xe mới.

### 1.3. Các API client có sẵn nhưng chưa gắn vào UI chính

Trong `api/client.ts`, frontend còn định nghĩa sẵn một số hàm API nhưng hiện tại UI chính không dùng trực tiếp:

- `healthCheck()`
- `cancelRide()`
- `getUserProfile()`
- `getRideHistory()`

## 2. Contract hiện tại giữa frontend và backend

Frontend hiện đang gọi base URL từ biến môi trường `REACT_APP_API_URL`, mặc định là `http://localhost:8080`.

Điểm cần lưu ý: trong code frontend, các endpoint đang được gọi theo prefix `/api/...`, nhưng ride service backend hiện tại expose route gốc là `/rides/...`.

### 2.1. Mapping frontend so với backend ride service thật

| Frontend đang gọi | Backend hiện có | Ghi chú |
| --- | --- | --- |
| `POST /api/rides` | `POST /rides` | Tạo chuyến xe |
| `GET /api/rides/{rideId}` | `GET /rides/{id}` | Lấy trạng thái/chuyến |
| `POST /api/rides/{rideId}/cancel` | `POST /rides/trips/{id}/cancel` | Frontend và backend đang lệch path |
| `GET /api/trips/ride/{rideId}` | Chưa thấy route tương ứng | Frontend tracking đang phụ thuộc endpoint này |
| `POST /api/trips/{tripId}/rating` | Chưa thấy route tương ứng | Backend hiện tại chưa có API rating |
| `GET /health` | `GET /rides/health` | Frontend health check cũng đang lệch path |
| `GET /api/users/profile` | Chưa thấy route tương ứng | Chỉ là hàm client, chưa được UI chính dùng |
| `GET /api/rides/history?limit=10` | `GET /rides/passenger/{id}` hoặc `GET /rides/driver/{id}` | Frontend đang dùng kiểu history khác với backend hiện tại |

## 3. Flow Postman để test MVP theo business logic chính

Mục tiêu của flow này là kiểm tra vòng đời chuyến đi từ lúc tạo ride đến lúc hoàn tất hoặc hủy.

### 3.1. Chuẩn bị biến môi trường Postman

Nên tạo collection variables như sau:

- `baseUrl`: URL của ride service hoặc API gateway đang expose ride service
- `passengerId`: ví dụ `pax-123`
- `driverId`: ví dụ `drv-456`
- `pickupLat`: `10.7769`
- `pickupLng`: `106.7009`
- `dropoffLat`: `10.8141`
- `dropoffLng`: `106.6269`

Nếu gọi thẳng ride service thì base path thực tế đang là `/rides/...`.

### 3.2. Flow happy path đề xuất để demo

#### Bước 1: Health check

`GET {{baseUrl}}/rides/health`

Kỳ vọng:

- HTTP `200`
- Body trả về kiểu:

```json
{
  "status": "healthy",
  "service": "ride-service"
}
```

#### Bước 2: Tạo chuyến xe

`POST {{baseUrl}}/rides`

Body:

```json
{
  "passengerId": "{{passengerId}}",
  "pickupLat": {{pickupLat}},
  "pickupLng": {{pickupLng}},
  "dropoffLat": {{dropoffLat}},
  "dropoffLng": {{dropoffLng}}
}
```

Kỳ vọng:

- HTTP `201`
- Response có `id`, `status = REQUESTED`, `estimatedFare`, `createdAt`

#### Bước 3: Lấy lại thông tin ride vừa tạo

`GET {{baseUrl}}/rides/{{rideId}}`

Kỳ vọng:

- HTTP `200`
- `status` vẫn là `REQUESTED`
- `passengerId` đúng
- tọa độ đúng

#### Bước 4: Gán tài xế cho ride

`PUT {{baseUrl}}/rides/{{rideId}}/status`

Body:

```json
{
  "status": "ASSIGNED",
  "driverId": "{{driverId}}"
}
```

Kỳ vọng:

- HTTP `200`
- `status = ASSIGNED`
- `driverId` được set

#### Bước 5: Bắt đầu chuyến đi

`POST {{baseUrl}}/rides/trips/{{rideId}}/start`

Kỳ vọng:

- HTTP `200`
- `status = IN_PROGRESS`
- `startedAt` được set

#### Bước 6: Hoàn tất chuyến đi

`POST {{baseUrl}}/rides/trips/{{rideId}}/complete`

Kỳ vọng:

- HTTP `200`
- `status = COMPLETED`
- `completedAt` được set

#### Bước 7: Verify cuối cùng

`GET {{baseUrl}}/rides/{{rideId}}`

Kỳ vọng:

- `status = COMPLETED`
- có đủ `startedAt` và `completedAt`

### 3.3. Flow hủy chuyến để test nhánh phụ

Để kiểm tra nhánh hủy, tạo một ride mới rồi gọi:

`POST {{baseUrl}}/rides/trips/{{rideId}}/cancel`

Kỳ vọng:

- Hủy được khi ride còn `REQUESTED` hoặc `ASSIGNED`
- Không hủy được khi ride đã `COMPLETED` hoặc đã `CANCELLED`

Sau đó gọi lại:

`GET {{baseUrl}}/rides/{{rideId}}`

Kỳ vọng:

- `status = CANCELLED`

## 4. Kết luận về mức độ hoàn chỉnh của MVP demo

### Đủ để demo phần nào

- Demo được luồng đặt xe cơ bản.
- Demo được vòng đời ride ở backend: tạo, gán tài xế, bắt đầu, hoàn tất, hủy.
- Demo được event-driven flow ở mức service nội bộ nếu Redis Stream và các consumer đang chạy.

### Chưa đủ nếu chỉ nhìn frontend hiện tại

- Frontend chưa có UI cho assign driver, start trip, complete trip.
- Frontend đang gọi một số endpoint `/api/trips/...` nhưng backend ride service hiện tại chưa expose các route đó.
- Frontend có feedback/rating UI, nhưng backend tương ứng chưa thấy API rating.
- Nếu muốn demo end-to-end trơn tru, cần thống nhất lại contract giữa frontend và backend, hoặc có API gateway rewrite path rõ ràng.

### Kết luận ngắn

Frontend hiện tại mới cover tốt phần rider-facing: đặt xe và xem tracking. Để chứng minh MVP đã hoàn chỉnh cho main business logic, bạn cần test đầy đủ ride service lifecycle ở Postman và xác nhận contract route giữa frontend với backend đã khớp.
