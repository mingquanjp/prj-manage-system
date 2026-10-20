export type TaskStatus =
  | "PENDING_ASSIGNMENT"
  | "ASSIGNED"
  | "IN_PROGRESS"
  | "PAUSED"
  | "PENDING_REVIEW"
  | "DONE"
  | "CLOSED"
  | "CANCELLED"
export type Project = {
  id: string
  code: string
  name: string
  description: string
  status: "PLANNING" | "IN_PROGRESS" | "COMPLETED" | "CLOSED" | "CANCELLED"
  pmId: string
  ptcmId?: string
  start: string
  end: string
  priority: "Thấp" | "Trung bình" | "Cao" | "Khẩn cấp"
}
export type Requirement = {
  skill: string
  level: number
  mandatory: boolean
  weight: number
}
export type TaskGroup = {
  id: string
  projectId: string
  name: string
  description: string
}
export type Task = {
  id: string
  projectId: string
  code: string
  name: string
  group: string
  groupId: string
  description: string
  criteria: string
  estimate: number
  status: TaskStatus
  requirements: Requirement[]
  assigneeId?: string
  firstAssignedAt?: string
  priority: "Thấp" | "Trung bình" | "Cao" | "Khẩn cấp"
  start?: string
  end?: string
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
export type SystemRole = "ADMIN" | "HR" | "DIRECTOR"
export const demoUsers: { personId: string; systemRoles: SystemRole[] }[] = [
  { personId: "khoa", systemRoles: [] },
  { personId: "thanh", systemRoles: [] },
  { personId: "cuong", systemRoles: [] },
  { personId: "an", systemRoles: [] },
  { personId: "binh", systemRoles: [] },
  { personId: "dung", systemRoles: [] },
  { personId: "ha", systemRoles: ["DIRECTOR"] },
]
export const isDirector = (actorId: string) =>
  demoUsers.some(
    (u) => u.personId === actorId && u.systemRoles.includes("DIRECTOR"),
  )
export const canViewProject = (project: Project, actorId: string) =>
  isDirector(actorId) || project.pmId === actorId || project.ptcmId === actorId

// Projection minh họa từ lịch sử đóng góp chưa nghiệm thu; không phải cột/bảng DB mới.
export type OutputConflict = {
  projectId: string
  taskId: string
  personId: string
}
// Mẫu: Thanh từng góp đầu ra cho T007, hiện Cường nộp và chưa được nghiệm thu.
export const initialOutputConflicts: OutputConflict[] = [
  { projectId: "kho", taskId: "t7", personId: "thanh" },
]
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
  CLOSED: "Đã đóng",
  CANCELLED: "Đã hủy",
}
export const levels = ["", "Cơ bản", "Trung cấp", "Nâng cao", "Chuyên gia"]

export const people: Person[] = [
  {
    id: "ha",
    name: "Đặng Thu Hà",
    initials: "TH",
    position: "Giám đốc",
    seniority: "Senior",
    active: true,
    skills: {},
    backgroundLoad: 0,
  },
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
  { skill: "Node.js", level: 3, mandatory: true, weight: 1 },
  { skill: "PostgreSQL", level: 3, mandatory: false, weight: 1 },
]
const frontend: Requirement[] = [
  { skill: "React", level: 3, mandatory: true, weight: 1 },
  { skill: "TypeScript", level: 3, mandatory: false, weight: 1 },
]
export const skillCatalog = ["Node.js", "PostgreSQL", "React", "TypeScript"]
export const initialGroups: TaskGroup[] = projects.flatMap((project) => [
  {
    id: `${project.id}-backend`,
    projectId: project.id,
    name: "Backend",
    description: "API, dữ liệu và xử lý nghiệp vụ.",
  },
  {
    id: `${project.id}-frontend`,
    projectId: project.id,
    name: "Frontend",
    description: "Giao diện và tương tác người dùng.",
  },
  {
    id: `${project.id}-testing`,
    projectId: project.id,
    name: "Kiểm thử",
    description: "Kiểm tra các luồng nghiệp vụ trước bàn giao.",
  },
])
const taskFixtures: Omit<Task, "groupId" | "priority">[] = [
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
export const initialTasks: Task[] = taskFixtures.map((task) => ({
  ...task,
  groupId: `${task.projectId}-${task.group.toLowerCase()}`,
  priority: "Trung bình",
  firstAssignedAt:
    task.status !== "PENDING_ASSIGNMENT"
      ? "2026-10-16T09:00:00+07:00"
      : undefined,
}))

export type TaskInput = Pick<
  Task,
  | "code"
  | "name"
  | "groupId"
  | "description"
  | "criteria"
  | "estimate"
  | "priority"
  | "start"
  | "end"
  | "requirements"
>
export type FormErrors = Record<string, string>

export function isTaskFrozen(task: Task) {
  return Boolean(task.firstAssignedAt) || task.status !== "PENDING_ASSIGNMENT"
}

export function validateTaskInput(
  project: Project,
  groups: TaskGroup[],
  tasks: Task[],
  input: TaskInput,
  editingId?: string,
): FormErrors {
  const errors: FormErrors = {}
  if (!input.code.trim() || input.code.trim().length > 30)
    errors.code = "Nhập mã công việc, tối đa 30 ký tự."
  else if (
    tasks.some(
      (t) =>
        t.projectId === project.id &&
        t.id !== editingId &&
        t.code === input.code.trim(),
    )
  )
    errors.code = "Mã công việc đã được dùng trong dự án."
  if (!input.name.trim() || input.name.trim().length > 150)
    errors.name = "Nhập tên công việc, tối đa 150 ký tự."
  if (!groups.some((g) => g.id === input.groupId && g.projectId === project.id))
    errors.groupId = "Chọn nhóm thuộc dự án này."
  if (!input.description.trim())
    errors.description = "Mô tả phạm vi và đầu ra mong đợi."
  if (!input.criteria.trim()) errors.criteria = "Bổ sung tiêu chí nghiệm thu."
  if (!Number.isFinite(input.estimate) || input.estimate <= 0)
    errors.estimate = "Ước tính phải là số giờ lớn hơn 0."
  if (!["Thấp", "Trung bình", "Cao", "Khẩn cấp"].includes(input.priority))
    errors.priority = "Chọn mức ưu tiên hợp lệ."
  if (Boolean(input.start) !== Boolean(input.end))
    errors.dates = "Nhập đủ ngày bắt đầu và kết thúc, hoặc để trống cả hai."
  else if (input.start && input.end) {
    const validDate = (value: string) =>
      /^\d{4}-\d{2}-\d{2}$/.test(value) &&
      !Number.isNaN(Date.parse(value)) &&
      new Date(value).toISOString().slice(0, 10) === value
    if (
      !validDate(input.start) ||
      !validDate(input.end) ||
      input.start > input.end ||
      input.start < project.start ||
      input.end > project.end
    )
      errors.dates =
        "Ngày công việc phải đúng thứ tự và nằm trong kế hoạch dự án."
  }
  const previous = tasks.find(
    (t) => t.id === editingId && t.projectId === project.id,
  )
  // Gia hạn có lý do/lịch sử là luồng riêng chưa dựng; không ghi đè cặp ngày đã lưu.
  if (
    previous?.start &&
    previous.end &&
    (input.start !== previous.start || input.end !== previous.end)
  ) {
    errors.dates =
      "Giữ nguyên kế hoạch đã lưu. Luồng điều chỉnh ngày có lý do/lịch sử chưa dựng trong đợt này."
  }
  if (!input.requirements.length)
    errors.requirements = "Thêm ít nhất một yêu cầu kỹ năng."
  input.requirements.forEach((r, i) => {
    if (!skillCatalog.includes(r.skill))
      errors[`skill-${i}`] = "Chọn kỹ năng trong danh mục."
    else if (
      input.requirements.some((other, j) => j !== i && other.skill === r.skill)
    )
      errors[`skill-${i}`] = "Kỹ năng này đã có trong công việc."
    if (!Number.isInteger(r.level) || r.level < 1 || r.level > 4)
      errors[`level-${i}`] = "Chọn một trong bốn mức."
    // Chỉ kiểm tra dữ liệu số hữu hạn cho demo, chưa chốt miền trọng số nghiệp vụ.
    if (!Number.isFinite(r.weight))
      errors[`weight-${i}`] = "Nhập giá trị số minh họa."
  })
  return errors
}

export function saveTask(
  project: Project,
  actorId: string,
  groups: TaskGroup[],
  tasks: Task[],
  input: TaskInput,
  editingId?: string,
): Task[] {
  if (!canAssign(project, actorId))
    throw new Error(
      "Chỉ PTCM của dự án đang lập kế hoạch/triển khai được sửa cấu trúc.",
    )
  const existing = editingId
    ? tasks.find((t) => t.id === editingId && t.projectId === project.id)
    : undefined
  if (editingId && (!existing || isTaskFrozen(existing)))
    throw new Error(
      "Công việc đã từng phân công không được sửa lại đặc tả bằng biểu mẫu này.",
    )
  const errors = validateTaskInput(project, groups, tasks, input, editingId)
  if (Object.keys(errors).length)
    throw new Error(Object.values(errors).join(" "))
  const group = groups.find((g) => g.id === input.groupId)!
  const task: Task = {
    ...existing,
    ...structuredClone(input),
    id: existing?.id ?? crypto.randomUUID(),
    projectId: project.id,
    code: input.code.trim(),
    name: input.name.trim(),
    description: input.description.trim(),
    criteria: input.criteria.trim(),
    group: group.name,
    start: input.start || undefined,
    end: input.end || undefined,
    status: existing?.status ?? "PENDING_ASSIGNMENT",
  }
  return existing
    ? tasks.map((t) => (t.id === task.id ? task : t))
    : [...tasks, task]
}

export function saveGroup(
  project: Project,
  actorId: string,
  groups: TaskGroup[],
  tasks: Task[],
  input: Pick<TaskGroup, "name" | "description">,
  editingId?: string,
) {
  if (!canAssign(project, actorId))
    throw new Error(
      "Chỉ PTCM của dự án đang lập kế hoạch/triển khai được sửa nhóm.",
    )
  const name = input.name.trim()
  if (!name || name.length > 150)
    throw new Error("Nhập tên nhóm, tối đa 150 ký tự.")
  if (
    groups.some(
      (g) =>
        g.projectId === project.id && g.id !== editingId && g.name === name,
    )
  )
    throw new Error("Tên nhóm đã có trong dự án.")
  if (
    editingId &&
    !groups.some((g) => g.id === editingId && g.projectId === project.id)
  )
    throw new Error("Không tìm thấy nhóm trong dự án.")
  const group = {
    id: editingId ?? crypto.randomUUID(),
    projectId: project.id,
    name,
    description: input.description.trim(),
  }
  return {
    groups: editingId
      ? groups.map((g) => (g.id === editingId ? group : g))
      : [...groups, group],
    tasks: tasks.map((t) =>
      t.projectId === project.id && t.groupId === editingId
        ? { ...t, group: name }
        : t,
    ),
  }
}

export function getPerson(id?: string) {
  return people.find((p) => p.id === id)
}
export type ProjectInput = Pick<
  Project,
  "code" | "name" | "pmId" | "start" | "end" | "priority" | "description"
>
export const priorities: Project["priority"][] = [
  "Thấp",
  "Trung bình",
  "Cao",
  "Khẩn cấp",
]
export function validateProjectInput(
  all: Project[],
  input: ProjectInput,
): FormErrors {
  const errors: FormErrors = {}
  if (!input.code.trim() || input.code.trim().length > 30)
    errors.code = "Nhập mã dự án, tối đa 30 ký tự."
  else if (all.some((p) => p.code === input.code.trim()))
    errors.code = "Mã dự án đã tồn tại."
  if (!input.name.trim() || input.name.trim().length > 150)
    errors.name = "Nhập tên dự án, tối đa 150 ký tự."
  if (!getPerson(input.pmId)?.active)
    errors.pmId = "Chọn nhân viên đang hoạt động làm PM."
  const validDate = (value: string) =>
    /^\d{4}-\d{2}-\d{2}$/.test(value) &&
    Number.isFinite(Date.parse(value)) &&
    new Date(value).toISOString().slice(0, 10) === value
  if (!validDate(input.start)) errors.start = "Nhập ngày bắt đầu hợp lệ."
  if (!validDate(input.end)) errors.end = "Nhập hạn dự kiến hợp lệ."
  else if (!errors.start && input.end < input.start)
    errors.end = "Hạn dự kiến không trước ngày bắt đầu."
  if (!priorities.includes(input.priority))
    errors.priority = "Chọn mức ưu tiên."
  return errors
}
export function createProject(
  all: Project[],
  actorId: string,
  input: ProjectInput,
): Project[] {
  if (!isDirector(actorId) || !getPerson(actorId)?.active)
    throw new Error("Chỉ Giám đốc được tạo dự án.")
  const errors = validateProjectInput(all, input)
  if (Object.keys(errors).length) throw new Error(Object.values(errors)[0])
  return [
    ...all,
    {
      ...input,
      id: crypto.randomUUID(),
      code: input.code.trim(),
      name: input.name.trim(),
      description: input.description.trim(),
      status: "PLANNING",
    },
  ]
}
export function ptcmBlockReason(
  project: Project,
  personId: string,
  tasks: Task[],
  conflicts: OutputConflict[],
): string | undefined {
  if (!getPerson(personId)?.active) return "Nhân viên không còn hoạt động."
  if (
    tasks.some(
      (t) =>
        t.projectId === project.id &&
        t.assigneeId === personId &&
        ["ASSIGNED", "IN_PROGRESS", "PAUSED", "PENDING_REVIEW"].includes(
          t.status,
        ),
    )
  )
    return "Còn công việc đang được giao trong dự án. Cần xử lý phân công trước khi chỉ định PTCM."
  if (
    conflicts.some(
      (c) =>
        c.projectId === project.id &&
        c.personId === personId &&
        tasks.some(
          (t) =>
            t.id === c.taskId &&
            t.projectId === project.id &&
            !["DONE", "CLOSED", "CANCELLED"].includes(t.status),
        ),
    )
  )
    return "Còn đầu ra đã đóng góp cần nghiệm thu; không thể tự đánh giá kết quả của mình."
}
export function appointPtcm(
  project: Project,
  actorId: string,
  personId: string,
  tasks: Task[],
  conflicts: OutputConflict[],
): Project {
  if (
    project.pmId !== actorId ||
    !getPerson(actorId)?.active ||
    !["PLANNING", "IN_PROGRESS"].includes(project.status)
  )
    throw new Error(
      "Chỉ PM của dự án đang lập kế hoạch hoặc triển khai được chỉ định PTCM.",
    )
  const reason = ptcmBlockReason(project, personId, tasks, conflicts)
  if (reason) throw new Error(reason)
  return { ...project, ptcmId: personId }
}
export function startProject(
  project: Project,
  actorId: string,
  tasks: Task[],
): Project {
  if (project.pmId !== actorId || !getPerson(actorId)?.active)
    throw new Error("Chỉ PM của dự án được bắt đầu triển khai.")
  if (project.status !== "PLANNING")
    throw new Error("Dự án không còn ở trạng thái lập kế hoạch.")
  if (!tasks.some((t) => t.projectId === project.id && t.status === "ASSIGNED"))
    throw new Error("Cần ít nhất một công việc đã phân công để bắt đầu dự án.")
  return { ...project, status: "IN_PROGRESS" }
}
export type ProjectEvent = {
  projectId: string
  actorId: string
  at: string
  action: string
  before?: Project
  after: Project
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
  const fixtureGroup = task.requirements.some((r) => r.skill === "React")
    ? "Frontend"
    : task.requirements.some((r) => r.skill === "Node.js")
      ? "Backend"
      : undefined
  const ids = fixtureGroup ? demoRanks[fixtureGroup] : []
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
      ? {
          ...task,
          status: "ASSIGNED",
          assigneeId: draft[task.id].personId,
          firstAssignedAt: task.firstAssignedAt ?? new Date().toISOString(),
        }
      : task,
  )
}
