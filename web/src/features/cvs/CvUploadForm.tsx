import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { validateCvFile } from "@/demo/employee-cvs"
import type { CvFileInput } from "@/models/employee-cvs"

export function CvUploadForm({
  employeeName,
  isOnBehalf,
  onSave,
  onClose,
}: {
  employeeName: string
  isOnBehalf: boolean
  onSave: (file: CvFileInput) => void
  onClose: () => void
}) {
  const [file, setFile] = useState<CvFileInput | null>(null)
  const [error, setError] = useState("")
  const [discard, setDiscard] = useState(false)
  const requestClose = () => (file ? setDiscard(true) : onClose())
  return (
    <>
      <Dialog
        open
        onOpenChange={(open) => {
          if (!open) requestClose()
        }}
        title={isOnBehalf ? "Bổ sung CV thay nhân viên" : "Bổ sung CV"}
        description={`Hồ sơ của ${employeeName}. ${isOnBehalf ? "Bạn được ghi nhận là người bổ sung; nhân viên tự chọn bản hiện tại." : "Thêm phiên bản mới không thay CV hiện tại."}`}
      >
        <form
          className="space-y-5"
          onSubmit={(event) => {
            event.preventDefault()
            try {
              if (!file) throw new Error("Chọn tệp CV trước khi bổ sung.")
              validateCvFile(file)
              onSave(file)
              onClose()
            } catch (err) {
              setError(err instanceof Error ? err.message : "Chưa thể bổ sung CV.")
            }
          }}
        >
          <div>
            <label className="field-label mb-2 block" htmlFor="cv-file">
              Tệp CV <span aria-hidden="true">*</span>
            </label>
            <input
              id="cv-file"
              type="file"
              accept=".pdf,.docx"
              required
              aria-describedby="cv-file-hint"
              className="field w-full min-w-0 text-sm file:mr-3 file:cursor-pointer"
              onChange={(event) => {
                const selected = event.target.files?.[0]
                setFile(
                  selected
                    ? { name: selected.name, size: selected.size, type: selected.type }
                    : null,
                )
                setError("")
              }}
            />
            <p id="cv-file-hint" className="mt-2 text-xs leading-5 text-muted-foreground">
              PDF có văn bản hoặc DOCX, tối đa 10 MiB. Không dùng PDF scan hoặc có mật khẩu.
            </p>
          </div>
          <p className="rounded-lg bg-accent p-3 text-xs leading-5 text-muted-foreground">
            Bản demo chỉ ghi tên, loại và dung lượng tệp trong phiên. Không đọc nội dung, tải lên
            server hoặc lưu tệp thật. Kiểm tra PDF có văn bản/mật khẩu và nội dung DOCX chưa triển
            khai.
          </p>
          {error && (
            <p role="alert" className="text-sm text-warning">
              {error}
            </p>
          )}
          <div className="flex flex-wrap justify-end gap-2 border-t border-border pt-4">
            <Button variant="outline" onClick={requestClose}>
              Hủy
            </Button>
            <Button type="submit" disabled={!file}>
              Bổ sung CV
            </Button>
          </div>
        </form>
      </Dialog>
      <Dialog
        open={discard}
        onOpenChange={setDiscard}
        title="Bỏ tệp đã chọn?"
        description="Phiên bản CV chưa được bổ sung. CV hiện tại giữ nguyên."
      >
        <div className="flex flex-wrap justify-end gap-2">
          <Button variant="outline" onClick={() => setDiscard(false)}>
            Tiếp tục
          </Button>
          <Button onClick={onClose}>Bỏ lựa chọn</Button>
        </div>
      </Dialog>
    </>
  )
}
