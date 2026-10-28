import { useIdentity } from "@/components/identity"
import { useSkillCatalog } from "@/components/skill-catalog"
import { useState } from "react"
import { usePeople } from "@/components/people-directory"
import { Avatar, EmptyState, SearchInput } from "@/components/shared"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { levels } from "@/models/workspace-labels"
import type { Project, Task } from "@/models/workspace"
import { canLookupPersonnel, personnelProfiles } from "@/demo/personnel"

export function PersonnelDirectory({
  projects,
  tasks,
  actorId,
}: {
  projects: Project[]
  tasks: Task[]
  actorId: string
}) {
  const identity = useIdentity()

  const { catalog } = useSkillCatalog()
  const { people } = usePeople()
  const [query, setQuery] = useState("")
  const [selectedId, setSelectedId] = useState<string>()
  if (!canLookupPersonnel(projects, actorId, identity)) {
    return (
      <EmptyState
        title="Bạn không có quyền tra cứu nhân sự"
        detail="Chức năng dành cho người đang được giao trách nhiệm PM hoặc PTCM trong dự án."
      />
    )
  }
  const profiles = personnelProfiles(projects, tasks, actorId, query, people, catalog, identity)
  const selected = profiles.find((person) => person.id === selectedId)

  return (
    <>
      <div className="mb-8">
        <p className="eyebrow">KHÔNG GIAN LÀM VIỆC</p>
        <h1 className="mt-2 text-2xl font-semibold text-balance">Tra cứu nhân sự</h1>
        <p className="mt-2 max-w-2xl text-sm text-muted-foreground text-pretty">
          Tìm hiểu chuyên môn và tải công việc để hỗ trợ điều phối dự án. Đây là góc nhìn nghề
          nghiệp, không phải hồ sơ hành chính.
        </p>
      </div>
      <div className="mb-5 flex flex-wrap items-end justify-between gap-4">
        <div className="w-full max-w-md">
          <label className="field-label mb-2 block" htmlFor="personnel-search">
            Tìm nhân sự
          </label>
          <SearchInput
            id="personnel-search"
            value={query}
            placeholder="Tìm tên, vị trí hoặc kỹ năng…"
            onChange={(value) => {
              setQuery(value)
              setSelectedId(undefined)
            }}
          />
        </div>
        <p role="status" className="text-sm text-muted-foreground tabular-nums">
          {profiles.length} nhân sự
        </p>
      </div>
      <div className="overflow-hidden rounded-xl border border-border bg-white">
        {profiles.length ? (
          <div className="overflow-x-auto">
            <table className="w-full min-w-[680px] text-left text-sm">
              <thead className="table-heading">
                <tr>
                  <th scope="col" className="px-5 py-3">
                    Nhân sự
                  </th>
                  <th scope="col" className="px-5 py-3">
                    Kỹ năng chính thức
                  </th>
                  <th scope="col" className="px-5 py-3 text-right">
                    Tải hiện tại
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {profiles.map((person) => (
                  <tr key={person.id} className="hover:bg-accent/40">
                    <td className="px-5 py-4">
                      <div className="flex items-center gap-3">
                        <Avatar initials={person.initials} />
                        <div>
                          <button
                            type="button"
                            className="text-left font-medium hover:underline"
                            aria-label={`Xem ${person.name}`}
                            onClick={() => setSelectedId(person.id)}
                          >
                            {person.name}
                          </button>
                          <p className="mt-1 text-xs text-muted-foreground">
                            {person.position} · {person.seniority}
                          </p>
                        </div>
                      </div>
                    </td>
                    <td className="max-w-64 px-5 py-4">
                      <div className="flex flex-wrap gap-1.5">
                        {person.skills.length ? (
                          person.skills.map((skill) => (
                            <span key={skill.name} className="rounded bg-accent px-2 py-1 text-xs">
                              {skill.name}
                            </span>
                          ))
                        ) : (
                          <span className="text-muted-foreground">Chưa có dữ liệu</span>
                        )}
                      </div>
                    </td>
                    <td className="px-5 py-4 text-right tabular-nums whitespace-nowrap">
                      {person.currentLoad} giờ
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <>
            <EmptyState
              title="Không tìm thấy nhân sự"
              detail="Thử tên, vị trí hoặc kỹ năng khác."
            />
            <div className="pb-6 text-center">
              <Button variant="outline" onClick={() => setQuery("")}>
                Xóa tìm kiếm
              </Button>
            </div>
          </>
        )}
      </div>
      <p className="mt-4 max-w-3xl text-xs leading-5 text-muted-foreground text-pretty">
        Tải là tổng giờ ước tính của công việc còn mở, gồm tải nền minh họa từ các dự án khác. Không
        phải số giờ đã làm hay lịch rảnh; không bao gồm lựa chọn phân công chưa xác nhận.
      </p>
      <Dialog
        side
        open={Boolean(selected)}
        onOpenChange={(open) => {
          if (!open) setSelectedId(undefined)
        }}
        title={selected?.name ?? "Thông tin nghề nghiệp"}
        description="Thông tin nghề nghiệp được phép xem. Dữ liệu trong phiên mô phỏng."
      >
        {selected && (
          <div className="space-y-7">
            <section aria-label="Thông tin công tác">
              <div className="flex items-center gap-3">
                <Avatar initials={selected.initials} />
                <div>
                  <h2 className="section-title">Thông tin công tác</h2>
                  <p className="mt-1 text-sm">
                    {selected.position} · {selected.seniority}
                  </p>
                </div>
              </div>
            </section>
            <section>
              <h2 className="section-title mb-3">Kỹ năng chính thức</h2>
              {selected.skills.length ? (
                <dl className="divide-y divide-border rounded-lg border border-border px-4">
                  {selected.skills.map((skill) => (
                    <div key={skill.name} className="flex justify-between gap-4 py-3 text-sm">
                      <dt className="font-medium">{skill.name}</dt>
                      <dd>{levels[skill.level]}</dd>
                    </div>
                  ))}
                </dl>
              ) : (
                <p className="text-sm text-muted-foreground">Chưa có dữ liệu kỹ năng chính thức.</p>
              )}
              <p className="mt-2 text-xs text-muted-foreground">
                Bốn mức: Cơ bản, Trung cấp, Nâng cao, Chuyên gia.
              </p>
            </section>
            <section>
              <h2 className="section-title mb-3">Tải công việc hiện tại</h2>
              <p className="text-2xl font-semibold tabular-nums">
                {selected.currentLoad} <span className="text-sm font-normal">giờ ước tính</span>
              </p>
              <p className="mt-2 text-sm leading-6 text-muted-foreground text-pretty">
                Tải tổng hợp xuyên dự án; không hiển thị công việc hoặc nhật ký của dự án khác. Số
                giờ đã ghi không được trừ khỏi tải ước tính.
              </p>
            </section>
            <section>
              <h2 className="section-title mb-3">Kinh nghiệm và kết quả thực hiện</h2>
              <p className="text-sm leading-6 text-muted-foreground text-pretty">
                Chưa có dữ liệu tổng hợp được xác minh trong bản mô phỏng. Không đồng nghĩa nhân sự
                chưa có kinh nghiệm; việc được phân công chưa được tính thành kết quả thực hiện.
              </p>
            </section>
          </div>
        )}
      </Dialog>
    </>
  )
}
