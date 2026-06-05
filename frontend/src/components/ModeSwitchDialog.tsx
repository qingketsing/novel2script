import * as Dialog from "@radix-ui/react-dialog";

type ModeSwitchDialogProps = {
  open: boolean;
  onCancel: () => void;
  onConfirm: () => void;
};

export function ModeSwitchDialog({ open, onCancel, onConfirm }: ModeSwitchDialogProps) {
  return (
    <Dialog.Root open={open} onOpenChange={(nextOpen) => (!nextOpen ? onCancel() : undefined)}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-zinc-950/30" />
        <Dialog.Content className="fixed left-1/2 top-1/2 z-50 w-[calc(100vw-2rem)] max-w-md -translate-x-1/2 -translate-y-1/2 rounded-lg border border-zinc-200 bg-white p-5 shadow-xl shadow-zinc-950/10">
          <Dialog.Title className="text-lg font-semibold text-zinc-950">切换到文本输入？</Dialog.Title>
          <Dialog.Description className="mt-2 text-sm leading-6 text-zinc-600">
            你当前已选择文件。继续编辑文本会清空已选文件，并改用文本转换接口。
          </Dialog.Description>
          <div className="mt-5 flex justify-end gap-2">
            <button className="secondary-button" type="button" onClick={onCancel}>
              取消
            </button>
            <button className="primary-button w-auto px-4 py-2" type="button" onClick={onConfirm}>
              切换
            </button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
