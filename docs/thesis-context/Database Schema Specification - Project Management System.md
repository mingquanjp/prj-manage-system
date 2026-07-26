# **Tài liệu Đặc tả Cơ sở dữ liệu (Database Schema Specification)**

Tài liệu mô tả chi tiết các bảng, kiểu dữ liệu, ràng buộc và ý nghĩa của từng trường trong hệ thống quản lý dự án, phân bổ công việc và phân tích kỹ năng nhân sự.

## **1\. Các kiểu dữ liệu tự định nghĩa (ENUM Types)**

Hệ thống sử dụng các kiểu ENUM sau để chuẩn hóa dữ liệu trạng thái và phân loại:

* **UserSystemRole:** ADMIN, HR, EMPLOYEE (Vai trò hệ thống)  
* **UserStatus:** ACTIVE, LOCKED, DISABLED (Trạng thái tài khoản)  
* **EmploymentStatus:** ACTIVE, INACTIVE, RESIGNED (Trạng thái làm việc của nhân viên)  
* **PriorityLevel:** LOW, MEDIUM, HIGH, URGENT (Mức độ ưu tiên)  
* **ProjectStatus:** PLANNING, IN\_PROGRESS, COMPLETED, CLOSED, CANCELLED (Trạng thái dự án)  
* **TaskStatus:** PENDING\_ASSIGNMENT, ASSIGNED, IN\_PROGRESS, DONE, CLOSED (Trạng thái công việc)  
* **SkillStatus:** ACTIVE, INACTIVE (Trạng thái kỹ năng)  
* **ProficiencyLevel:** BASIC, INTERMEDIATE, ADVANCED, EXPERT (Mức độ thành thạo)  
* **JoinSource:** MANUAL, FROM\_ASSIGNMENT (Nguồn thêm vào dự án)  
* **AssignmentType:** AI, MANUAL (Phương thức phân công)  
* **CvStatus:** ACTIVE, ARCHIVED, REPLACED (Trạng thái hồ sơ CV)  
* **ExtractionStatus:** PENDING\_REVIEW, APPROVED, REJECTED (Trạng thái trích xuất CV)  
* **DecisionStatus:** PENDING, APPROVED, REJECTED (Trạng thái quyết định phê duyệt)  
* **ChangeSource:** MANUAL, SYSTEM (Nguồn thay đổi trạng thái)

## **2\. Chi tiết cấu trúc các bảng (Tables)**

### **2.1. Bảng Users (Tài khoản người dùng)**

| Cột | Kiểu dữ liệu | Ràng buộc | Mô tả |
| :---- | :---- | :---- | :---- |
| user\_id | INT | PK, Identity | Khóa chính định danh tài khoản. |
| username | varchar(50) | UNIQUE, NOT NULL | Tên đăng nhập hệ thống. |
| password\_hash | varchar(255) | NOT NULL | Mật khẩu đã được mã hóa. |
| system\_role | UserSystemRole | NOT NULL | Quyền hạn trên hệ thống. |
| status | UserStatus | NOT NULL, DEFAULT 'ACTIVE' | Trạng thái hiện tại của tài khoản. |
| created\_at | datetime | NOT NULL | Thời điểm tạo tài khoản. |
| updated\_at | datetime | NOT NULL | Thời điểm cập nhật gần nhất. |

### **2.2. Bảng Departments (Phòng ban)**

| Cột | Kiểu dữ liệu | Ràng buộc | Mô tả |
| :---- | :---- | :---- | :---- |
| department\_id | INT | PK, Identity | Khóa chính phòng ban. |
| department\_name | varchar(100) | UNIQUE, NOT NULL | Tên phòng ban. |
| created\_at | datetime | NOT NULL | Thời gian tạo. |
| updated\_at | datetime | NOT NULL | Thời gian cập nhật. |

### **2.3. Bảng Positions (Vị trí chức danh)**

| Cột | Kiểu dữ liệu | Ràng buộc | Mô tả |
| :---- | :---- | :---- | :---- |
| position\_id | INT | PK, Identity | Khóa chính định danh chức danh. |
| position\_name | varchar(100) | UNIQUE, NOT NULL | Tên chức danh (VD: Developer, PM). |
| created\_at | datetime | NOT NULL | Thời gian tạo. |
| updated\_at | datetime | NOT NULL | Thời gian cập nhật. |

### **2.4. Bảng Employees (Nhân viên)**

| Cột | Kiểu dữ liệu | Ràng buộc | Mô tả |
| :---- | :---- | :---- | :---- |
| employee\_id | INT | PK, Identity | Khóa chính nhân viên. |
| user\_id | int | FK, UNIQUE, NOT NULL | Liên kết tài khoản người dùng (1-1). |
| employee\_code | varchar(30) | UNIQUE, NOT NULL | Mã định danh nội bộ của nhân sự. |
| fullname | varchar(100) | NOT NULL | Họ và tên đầy đủ. |
| email | varchar(100) | UNIQUE, NOT NULL | Email làm việc liên hệ. |
| phone | varchar(20) | NULL | Số điện thoại cá nhân. |
| department\_id | int | FK, NULL | Khóa ngoại tới phòng ban trực thuộc. |
| position\_id | int | FK, NOT NULL | Khóa ngoại tới chức danh. |
| seniority\_level | varchar(30) | NULL | Cấp bậc kinh nghiệm (Junior, Senior...). |
| employment\_status | EmploymentStatus | NOT NULL, DEFAULT 'ACTIVE' | Trạng thái làm việc tại công ty. |

### **2.5. Bảng SkillCategories (Danh mục kỹ năng)**

| Cột | Kiểu dữ liệu | Ràng buộc | Mô tả |
| :---- | :---- | :---- | :---- |
| category\_id | INT | PK, Identity | Khóa chính danh mục. |
| category\_name | varchar(100) | UNIQUE, NOT NULL | Tên nhóm kỹ năng (Backend, Soft skill...). |

### **2.6. Bảng Skills (Kỹ năng chi tiết)**

| Cột | Kiểu dữ liệu | Ràng buộc | Mô tả |
| :---- | :---- | :---- | :---- |
| skill\_id | INT | PK, Identity | Khóa chính định danh kỹ năng. |
| skill\_name | varchar(100) | UNIQUE, NOT NULL | Tên kỹ năng (VD: Java, React). |
| category\_id | int | FK, NOT NULL | Thuộc nhóm danh mục nào. |
| status | SkillStatus | NOT NULL, DEFAULT 'ACTIVE' | Tình trạng sử dụng kỹ năng. |

### **2.7. Bảng EmployeeSkills (Hồ sơ kỹ năng nhân viên)**

| Cột | Kiểu dữ liệu | Ràng buộc | Mô tả |
| :---- | :---- | :---- | :---- |
| employee\_skill\_id | INT | PK, Identity | Định danh bản ghi năng lực. |
| employee\_id | int | FK, NOT NULL | Người sở hữu kỹ năng. |
| skill\_id | int | FK, NOT NULL | Loại kỹ năng. |
| proficiency\_level | ProficiencyLevel | NOT NULL | Mức độ thành thạo đánh giá. |
| years\_experience | decimal(4,1) | NULL | Số năm kinh nghiệm cụ thể. |
| note | text | NULL | Ghi chú thêm. |

### **2.8. Bảng Projects (Dự án)**

| Cột | Kiểu dữ liệu | Ràng buộc | Mô tả |
| :---- | :---- | :---- | :---- |
| project\_id | INT | PK, Identity | Khóa chính định danh dự án. |
| project\_code | varchar(30) | UNIQUE, NOT NULL | Mã quản lý dự án. |
| project\_name | varchar(150) | NOT NULL | Tên dự án. |
| description | text | NULL | Mô tả phạm vi và mục tiêu. |
| start\_date | date | NOT NULL | Ngày bắt đầu. |
| planned\_end\_date | date | NOT NULL | Ngày dự kiến kết thúc. |
| priority | PriorityLevel | NOT NULL, DEFAULT 'MEDIUM' | Mức độ ưu tiên. |
| status | ProjectStatus | NOT NULL, DEFAULT 'PLANNING' | Tiến độ và trạng thái vòng đời. |
| pm\_employee\_id | int | FK, NOT NULL | Quản trị viên dự án (PM). |

### **2.9. Bảng ProjectMembers (Thành viên dự án)**

| Cột | Kiểu dữ liệu | Ràng buộc | Mô tả |
| :---- | :---- | :---- | :---- |
| project\_member\_id | INT | PK, Identity | Khóa định danh thành viên trong dự án. |
| project\_id | int | FK, NOT NULL | Dự án được phân bổ. |
| employee\_id | int | FK, NOT NULL | Nhân viên tham gia. |
| join\_source | JoinSource | NOT NULL | Cách thức được thêm vào dự án. |
| joined\_at | datetime | NOT NULL | Thời gian bắt đầu tham gia. |

### **2.10. Bảng Tasks (Nhiệm vụ / Công việc)**

| Cột | Kiểu dữ liệu | Ràng buộc | Mô tả |
| :---- | :---- | :---- | :---- |
| task\_id | INT | PK, Identity | Khóa chính định danh task. |
| project\_id | int | FK, NOT NULL | Thuộc dự án nào. |
| parent\_task\_id | int | FK, NULL | Công việc cha (Self-referencing). |
| current\_assignee\_id | int | FK, NULL | Người được giao việc hiện tại. |
| task\_code | varchar(30) | UNIQUE, NOT NULL | Mã quản lý công việc. |
| task\_name | varchar(150) | NOT NULL | Tiêu đề công việc. |
| task\_type | varchar(30) | NOT NULL, DEFAULT 'CHILD' | Loại hình công việc. |
| estimate\_hours | decimal(10,2) | NOT NULL | Thời lượng dự kiến hoàn thành. |
| actual\_hours | decimal(10,2) | NOT NULL, DEFAULT 0 | Thời gian thực tế đã log. |
| status | TaskStatus | NOT NULL, DEFAULT 'PENDING\_ASSIGNMENT' | Trạng thái tiến độ thực thi. |

### **2.11. Bảng TaskRequirements (Yêu cầu kỹ năng cho Task)**

| Cột | Kiểu dữ liệu | Mô tả |
| :---- | :---- | :---- |
| task\_req\_id | INT | Khóa chính định danh yêu cầu. |
| task\_id / skill\_id | int | Các khóa ngoại liên kết Công việc và Kỹ năng cần có. |
| required\_level | ProficiencyLevel | Mức độ am hiểu tối thiểu cần thiết để làm task. |
| weight | decimal(5,2) | Trọng số ưu tiên của kỹ năng trong task này. |
| is\_mandatory | boolean | Yêu cầu bắt buộc (True) hay tùy chọn (False). |

### **2.12. Bảng Assignments (Phân công nhiệm vụ)**

| Cột | Kiểu dữ liệu | Mô tả |
| :---- | :---- | :---- |
| assignment\_id | INT | Khóa chính lịch sử phân công. |
| task\_id / employee\_id | int | Khóa ngoại Task được giao và Người nhận. |
| assigned\_by\_employee\_id | int | Người quản lý thực hiện phân công. |
| assignment\_type | AssignmentType | Phân công theo gợi ý AI hay thủ công (MANUAL). |
| matching\_score | decimal(5,2) | Điểm độ tương thích kỹ năng (nếu từ AI). |

### **2.13. Bảng WorkLogs (Ghi nhận thời gian làm việc)**

| Cột | Kiểu dữ liệu | Mô tả |
| :---- | :---- | :---- |
| timelog\_id | INT | Định danh bản log. |
| task\_id / employee\_id | int | Khóa ngoại Công việc và Người làm. |
| work\_date | date | Ngày thực hiện công việc. |
| hours\_spent | decimal(10,2) | Thời lượng (số giờ) tiêu tốn. |
| work\_note | text | Chi tiết nội dung đã hoàn thiện. |

### **2.14. Bảng EmployeeCVs (Hồ sơ CV của nhân viên)**

| Cột | Kiểu dữ liệu | Mô tả |
| :---- | :---- | :---- |
| cv\_id | INT | Định danh tệp CV. |
| storage\_provider / storage\_key | varchar | Nơi lưu trữ file và Khóa truy cập (AWS S3, Google Cloud...). |
| file\_url / original\_file\_name | text / varchar(255) | Đường dẫn tải xuống và tên file gốc. |
| status | CvStatus | Trạng thái hiệu lực của CV hiện tại. |

### **2.15. Hệ thống Bảng bổ trợ khác (Phân tích, Lịch sử, Bình luận)**

* **TaskComments:** Lưu trữ các trao đổi, bình luận của user trong quá trình thực hiện một Task cụ thể.  
* **CVExtractions:** Cấu trúc để tiếp nhận văn bản thô, Markdown được trích xuất (parsing text) từ file CV (EmployeeCVs). Có quy trình chờ review từ quản lý.  
* **CVExtractedSkills:** Nhận diện các từ khóa kỹ năng từ CVExtractions. Bao gồm mức độ tự động gợi ý (suggested\_level), điểm tin cậy AI (confidence\_score) và yêu cầu HR duyệt (decision\_status).  
* **TaskStatusHistory:** Hệ thống Audit Trail, lưu lại dòng thời gian chuyển đổi trạng thái của các task (từ trạng thái cũ sang mới, do hệ thống hay thao tác tay).

*Ghi chú: Tất cả các bảng quan trọng đều áp dụng Index hóa (B-Tree, Unique Index) dựa trên yêu cầu truy vấn (User\_id, mã Code, Status, Time ranges) và ràng buộc khóa ngoại (Foreign Keys) để toàn vẹn hóa dữ liệu.*