export type TaskStatus = "PENDING_ASSIGNMENT" | "ASSIGNED" | "IN_PROGRESS" | "PAUSED" | "PENDING_REVIEW" | "DONE" | "CANCELLED"
export type Project = {
  id: string
  code: string
  name: string
  description: string
  status: "PLANNING" | "IN_PROGRESS"
  pmId: string
  ptcmId: string
  start: string
  end: string
  priority: "Cao" | "Trung bình"
}
export type Requirement = {
  skill: string
  level: number
  mandatory: boolean
}
export type Task = {
  id: string
  projectId: string
  code: string
  name: string
  group: string
  description: string
  criteria: string
  estimate: number
  status: TaskStatus
  requirements: Requirement[]
  assigneeId?: string
}
export type Person = {
  id: string
  name: string
  initials: string
  position: string
  seniority: string
  active: boolean
  skills: Record<string, number>
  backgroundLoad: number
}
export type Choice = {
  personId: string
  mode: "manual" | "recommended"
  rank?: number
  reason: string
  overrideReason: string
  hadRun: boolean
}
export type Draft = Record<string, Choice>

export const statusLabels: Record<TaskStatus, string> = {
  PENDING_ASSIGNMENT: "Chờ phân công",
  ASSIGNED: "Đã phân công",
  IN_PROGRESS: "Đang thực hiện",
  PAUSED: "Tạm dừng",
  PENDING_REVIEW: "Chờ nghiệm thu",
  DONE: "Đã nghiệm thu",
  CANCELLED: "Đã hủy",
}
export const levels = ["", "Cơ bản", "Trung cấp", "Nâng cao", "Chuyên gia"]

export const people: Person[] = [
  {
    id: "khoa",
    name: "Ngô Minh Khoa",
    initials: "MK",
    position: "Kỹ sư phần mềm",
    seniority: "Senior",
    active: true,
    skills: { "Node.js": 4, PostgreSQL: 3, React: 3 },
    backgroundLoad: 0,
  },
  {
    id: "thanh",
    name: "Phan Thị Thanh",
    initials: "PT",
    position: "Quản lý dự án",
    seniority: "Senior",
    active: true,
    skills: { React: 3 },
    backgroundLoad: 0,
  },
  {
    id: "an",
    name: "Nguyễn Văn An",
    initials: "VA",
    position: "Kỹ sư Backend",
    seniority: "Senior",
    active: true,
    skills: { "Node.js": 4, PostgreSQL: 3, TypeScript: 3 },
    backgroundLoad: 16,
  },
  {
    id: "binh",
    name: "Trần Thị Bình",
    initials: "TB",
    position: "Kỹ sư Frontend",
    seniority: "Middle",
    active: true,
    skills: { React: 3, TypeScript: 3, "Node.js": 3, PostgreSQL: 2 },
    backgroundLoad: 8,
  },
  {
    id: "dung",
    name: "Phạm Quang Dũng",
    initials: "QD",
    position: "Kỹ sư phần mềm",
    seniority: "Middle",
    active: true,
    skills: { "Node.js": 3, PostgreSQL: 3, React: 2, TypeScript: 2 },
    backgroundLoad: 12,
  },
  {
    id: "cuong",
    name: "Lê Minh Cường",
    initials: "MC",
    position: "Kỹ sư Frontend",
    seniority: "Senior",
    active: true,
    skills: { React: 4, TypeScript: 4, "Node.js": 2 },
    backgroundLoad: 20,
  },
]

export const projects: Project[] = [
  {
    id: "kho",
    code: "KHO-2026",
    name: "Hệ thống quản lý kho hàng",
    description:
      "Xây dựng hệ thống quản lý nhập, xuất và tồn kho nội bộ; tập trung vào dữ liệu chính xác và thao tác thuận tiện.",
    status: "IN_PROGRESS",
    pmId: "thanh",
    ptcmId: "khoa",
    start: "2026-10-01",
    end: "2026-12-31",
    priority: "Cao",
  },
  {
    id: "portal",
    code: "PORTAL-2026",
    name: "Cổng thông tin nhân sự",
    description: "Không gian tra cứu và quản lý thông tin nhân sự nội bộ.",
    status: "PLANNING",
    pmId: "thanh",
    ptcmId: "khoa",
    start: "2026-10-15",
    end: "2027-02-28",
    priority: "Trung bình",
  },
]

const backend: Requirement[] = [
  { skill: "Node.js", level: 3, mandatory: true },
  { skill: "PostgreSQL", level: 3, mandatory: false },
]
const frontend: Requirement[] = [
  { skill: "React", level: 3, mandatory: true },
  { skill: "TypeScript", level: 3, mandatory: false },
]
export const initialTasks: Task[] = [
  {
    id: "t1",
    projectId: "kho",
    code: "KHO-T001",
    name: "API quản lý sản phẩm",
    group: "Backend",
    estimate: 24,
    status: "PENDING_ASSIGNMENT",
    requirements: backend,
    description:
      "Xây dựng API thêm, sửa và tra cứu sản phẩm. Đầu ra gồm API, kiểm tra dữ liệu đầu vào và tài liệu endpoint.",
    criteria:
      "Không trùng mã sản phẩm; trả lỗi rõ ràng; có kiểm thử các thao tác chính.",
  },
  {
    id: "t2",
    projectId: "kho",
    code: "KHO-T002",
    name: "API nhập và xuất kho",
    group: "Backend",
    estimate: 32,
    status: "PENDING_ASSIGNMENT",
    requirements: backend,
    description:
      "Xây dựng API phiếu nhập và xuất kho, cập nhật số tồn trong cùng giao dịch.",
    criteria:
      "Không xuất vượt tồn; phiếu và số tồn nhất quán khi thành công hoặc thất bại.",
  },
  {
    id: "t3",
    projectId: "kho",
    code: "KHO-T003",
    name: "Màn hình tổng quan tồn kho",
    group: "Frontend",
    estimate: 28,
    status: "PENDING_ASSIGNMENT",
    requirements: frontend,
    description:
      "Xây dựng giao diện danh sách tồn kho, lọc theo sản phẩm và kho; có trạng thái rỗng và lỗi.",
    criteria:
      "Bộ lọc hoạt động; số liệu khớp nguồn dữ liệu; sử dụng được bằng bàn phím.",
  },
  {
    id: "t4",
    projectId: "kho",
    code: "KHO-T004",
    name: "Màn hình phiếu nhập kho",
    group: "Frontend",
    estimate: 20,
    status: "IN_PROGRESS",
    requirements: frontend,
    assigneeId: "binh",
    description: "Xây dựng form nhập kho và bảng các dòng sản phẩm.",
    criteria:
      "Kiểm tra số lượng dương; tổng các dòng chính xác; hủy không lưu dữ liệu.",
  },
  {
    id: "t5",
    projectId: "kho",
    code: "KHO-T005",
    name: "Cấu trúc dữ liệu sản phẩm",
    group: "Backend",
    estimate: 16,
    status: "DONE",
    requirements: backend,
    assigneeId: "an",
    description:
      "Thiết kế cấu trúc lưu trữ sản phẩm và kho; đầu ra gồm sơ đồ và mô tả ràng buộc.",
    criteria: "Quan hệ rõ ràng, có khóa duy nhất và kiểm tra dữ liệu bắt buộc.",
  },
  {
    id: "t6",
    projectId: "kho",
    code: "KHO-T006",
    name: "API báo cáo biến động tồn kho",
    group: "Backend",
    estimate: 16,
    status: "ASSIGNED",
    requirements: backend,
    assigneeId: "dung",
    description:
      "Xây dựng API báo cáo nhập, xuất và tồn theo khoảng thời gian.",
    criteria:
      "Tồn đầu + nhập − xuất = tồn cuối; hỗ trợ trường hợp không phát sinh.",
  },
  {
    id: "t7",
    projectId: "kho",
    code: "KHO-T007",
    name: "Bộ lọc và tra cứu sản phẩm",
    group: "Frontend",
    estimate: 12,
    status: "PENDING_REVIEW",
    requirements: frontend,
    assigneeId: "cuong",
    description: "Bổ sung tìm kiếm và lọc cho danh sách sản phẩm.",
    criteria:
      "Giữ điều kiện khi mở và đóng chi tiết; hiển thị rõ kết quả rỗng.",
  },
  {
    id: "p1",
    projectId: "portal",
    code: "PORTAL-T001",
    name: "Danh sách hồ sơ nhân viên",
    group: "Frontend",
    estimate: 24,
    status: "PENDING_ASSIGNMENT",
    requirements: frontend,
    description: "Xây dựng danh sách hồ sơ theo phạm vi được xem.",
    criteria: "Có tìm kiếm, trạng thái rỗng và giới hạn trường theo quyền.",
  },
  {
    id: "p2",
    projectId: "portal",
    code: "PORTAL-T002",
    name: "API tra cứu năng lực",
    group: "Backend",
    estimate: 20,
    status: "ASSIGNED",
    requirements: backend,
    assigneeId: "an",
    description: "API đọc kỹ năng chính thức và tải công việc tổng hợp.",
    criteria: "Không trả dữ liệu liên hệ hoặc chi tiết công việc ngoài quyền.",
  },
]

export function getPerson(id?: string) {
  return people.find((p) => p.id === id)
}
export function canAssign(project: Project, actorId: string) {
  return (
    actorId === project.ptcmId &&
    ["PLANNING", "IN_PROGRESS"].includes(project.status)
  )
}
export function isEligible(project: Project, person: Person) {
  return person.active && project.ptcmId !== person.id
}
export function skillGaps(task: Task, person: Person) {
  return task.requirements.filter(
    (r) => r.mandatory && (person.skills[r.skill] ?? 0) < r.level,
  )
}
export function loadFor(personId: string, tasks: Task[]) {
  return (
    (getPerson(personId)?.backgroundLoad ?? 0) +
    tasks
      .filter(
        (t) =>
          t.assigneeId === personId &&
          ["ASSIGNED", "IN_PROGRESS", "PAUSED", "PENDING_REVIEW"].includes(
            t.status,
          ),
      )
      .reduce((sum, t) => sum + t.estimate, 0)
  )
}
export function addedLoad(personId: string, draft: Draft, tasks: Task[]) {
  return Object.entries(draft)
    .filter(([, choice]) => choice.personId === personId)
    .reduce(
      (sum, [id]) => sum + (tasks.find((t) => t.id === id)?.estimate ?? 0),
      0,
    )
}
export function progress(tasks: Task[]) {
  const effective = tasks.filter((t) => t.status !== "CANCELLED")
  return effective.length
    ? Math.round(
        (effective.filter((t) => t.status === "DONE").length /
          effective.length) *
          100,
      )
    : null
}

// Dữ liệu minh họa bố cục, không phải công thức cuối hoặc xếp hạng thời gian thực.
const demoRanks = {
  Backend: ["an", "dung", "binh"],
  Frontend: ["cuong", "binh"],
}
export function recommendations(project: Project, task: Task) {
  const ids = demoRanks[(task.group as keyof typeof demoRanks)] ?? []
  return ids
    .map((id) => getPerson(id)!)
    .filter((p) => isEligible(project, p) && skillGaps(task, p).length === 0)
    .slice(0, 3)
    .map((person, index) => ({
      person,
      rank: index + 1,
      score: [94.6, 87.2, 81.8][index],
    }))
}

export function validateDraft(
  project: Project,
  actorId: string,
  tasks: Task[],
  draft: Draft,
): string[] {
  if (!canAssign(project, actorId))
    return ["Bạn không có trách nhiệm phân công tại dự án này."]
  if (!Object.keys(draft).length) return ["Chưa có lựa chọn trong bản nháp."]
  const errors: string[] = []
  for (const [id, choice] of Object.entries(draft)) {
    const task = tasks.find((t) => t.id === id && t.projectId === project.id)
    const person = getPerson(choice.personId)
    if (!task || task.status !== "PENDING_ASSIGNMENT") {
      errors.push("Có công việc không còn chờ phân công.")
      continue
    }
    if (!person || !isEligible(project, person)) {
      errors.push(`${task.code}: người được chọn không đủ điều kiện.`)
      continue
    }
    if (
      choice.mode === "recommended" &&
      !recommendations(project, task).some(
        (r) => r.person.id === person.id && r.rank === choice.rank,
      )
    )
      errors.push(`${task.code}: kết quả gợi ý không hợp lệ.`)
    if (
      ((choice.mode === "recommended" && choice.rank !== 1) ||
        (choice.mode === "manual" && choice.hadRun)) &&
      !choice.reason.trim()
    )
      errors.push(`${task.code}: bổ sung lý do lựa chọn.`)
    if (
      skillGaps(task, person).length &&
      (!choice.overrideReason.trim() || choice.mode !== "manual")
    )
      errors.push(`${task.code}: bổ sung lý do ngoại lệ kỹ năng.`)
  }
  return errors
}
export function commitDraft(
  project: Project,
  actorId: string,
  tasks: Task[],
  draft: Draft,
): Task[] {
  const errors = validateDraft(project, actorId, tasks, draft)
  if (errors.length) throw new Error(errors.join(" "))
  return tasks.map((task) =>
    draft[task.id]
      ? { ...task, status: "ASSIGNED", assigneeId: draft[task.id].personId }
      : task,
  )
}
