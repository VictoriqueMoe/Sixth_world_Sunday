import { type ChangeEvent, type DragEvent, type MouseEvent, useEffect, useRef, useState } from "react";
import { useSearchParams } from "react-router";
import { useQueryClient } from "@tanstack/react-query";
import { useAuth } from "../../hooks/useAuth";
import { isSiteStaff } from "../../utils/permissions";
import { useVaultBrowse } from "../../api/queries/vault";
import {
    useCreateVaultFolder,
    useDeleteVaultFile,
    useDeleteVaultFolder,
    useRenameVaultFile,
    useRenameVaultFolder,
    useSetVaultFileLocked,
    useSetVaultFolderLocked,
} from "../../api/mutations/vault";
import { vaultDownloadUrl, vaultUploadUrl, type VaultFile, type VaultFolder } from "../../api/endpoints";
import { ContextMenu, type ContextMenuItem } from "../../components/ContextMenu/ContextMenu";
import { useContextMenu } from "../../components/ContextMenu/useContextMenu";
import { Modal } from "../../components/Modal/Modal";
import { Button } from "../../components/Button/Button";
import { ChannelRail } from "../../components/layout/ChannelRail/ChannelRail";
import { useIsMobile } from "../../hooks/useIsMobile";
import { useNotifications } from "../../hooks/useNotifications";
import type { WSMessage } from "../../types/api";
import styles from "./FileBrowserPage.module.css";

type UploadStatus = "pending" | "uploading" | "done" | "error";

interface UploadTask {
    id: number;
    name: string;
    file: File;
    progress: number;
    status: UploadStatus;
}

interface RenameTarget {
    kind: "folder" | "file";
    id: string;
    name: string;
}

interface DeleteTarget {
    kind: "folder" | "file";
    id: string;
    name: string;
}

const UPLOAD_CONCURRENCY = 3;

function formatSize(bytes: number): string {
    if (bytes < 1024) {
        return `${bytes} B`;
    }

    const units = ["KB", "MB", "GB", "TB"];
    let value = bytes / 1024;
    let unit = 0;
    while (value >= 1024 && unit < units.length - 1) {
        value /= 1024;
        unit++;
    }

    return `${value.toFixed(1)} ${units[unit]}`;
}

function uploadStatLabel(task: UploadTask): string {
    if (task.status === "error") {
        return "failed";
    }
    if (task.status === "done") {
        return "✓";
    }
    return `${task.progress}%`;
}

function folderKey(id: string): string {
    return `folder:${id}`;
}

function fileKey(id: string): string {
    return `file:${id}`;
}

function errorFromXhr(xhr: XMLHttpRequest): string {
    const fallback = `Upload failed (${xhr.status})`;
    try {
        const body = JSON.parse(xhr.responseText) as { error?: string };
        return body.error ?? fallback;
    } catch {
        return fallback;
    }
}

function uploadOne(url: string, file: File, folderId: string | null, onProgress: (pct: number) => void): Promise<void> {
    return new Promise((resolve, reject) => {
        const form = new FormData();
        form.append("file", file);
        if (folderId) {
            form.append("folderId", folderId);
        }

        const xhr = new XMLHttpRequest();
        xhr.open("POST", url);
        xhr.withCredentials = true;

        xhr.upload.onprogress = event => {
            if (event.lengthComputable) {
                onProgress(Math.round((event.loaded / event.total) * 100));
            }
        };
        xhr.onload = () => {
            if (xhr.status >= 200 && xhr.status < 300) {
                resolve();
                return;
            }
            reject(new Error(errorFromXhr(xhr)));
        };
        xhr.onerror = () => reject(new Error("Upload failed"));

        xhr.send(form);
    });
}

export function FileBrowserPage() {
    const { user } = useAuth();
    const qc = useQueryClient();
    const isMobile = useIsMobile();
    const { addWSListener } = useNotifications();
    const [searchParams, setSearchParams] = useSearchParams();
    const folderId = searchParams.get("folder");

    const browse = useVaultBrowse(folderId);
    const { state: menuState, open: openMenu, close: closeMenu } = useContextMenu();

    const createFolder = useCreateVaultFolder();
    const renameFolder = useRenameVaultFolder();
    const deleteFolder = useDeleteVaultFolder();
    const lockFolder = useSetVaultFolderLocked();
    const renameFile = useRenameVaultFile();
    const deleteFile = useDeleteVaultFile();
    const lockFile = useSetVaultFileLocked();

    const fileInputRef = useRef<HTMLInputElement>(null);
    const uploadIdRef = useRef(0);
    const [dragActive, setDragActive] = useState(false);
    const [uploads, setUploads] = useState<UploadTask[]>([]);
    const [selected, setSelected] = useState<Set<string>>(new Set());
    const [createOpen, setCreateOpen] = useState(false);
    const [createName, setCreateName] = useState("");
    const [renameTarget, setRenameTarget] = useState<RenameTarget | null>(null);
    const [renameValue, setRenameValue] = useState("");
    const [deleteTarget, setDeleteTarget] = useState<DeleteTarget | null>(null);
    const [bulkDeleteOpen, setBulkDeleteOpen] = useState(false);
    const [actionError, setActionError] = useState<string | null>(null);

    const data = browse.data;
    const canManageLocks = data?.canManageLocks ?? false;
    const staff = isSiteStaff(user?.role);
    const myId = user?.id;

    const folders = data?.folders ?? [];
    const files = data?.files ?? [];
    const itemCount = folders.length + files.length;
    const isEmpty = !browse.loading && itemCount === 0;
    const crumbs = data?.breadcrumbs ?? [];
    const allSelected = itemCount > 0 && selected.size === itemCount;

    const uploadDone = uploads.filter(u => u.status === "done").length;
    const uploadFailed = uploads.filter(u => u.status === "error").length;
    const uploadActive = uploads.some(u => u.status === "pending" || u.status === "uploading");

    useEffect(() => {
        document.body.setAttribute("data-chat-page", "true");
        return () => {
            document.body.removeAttribute("data-chat-page");
        };
    }, []);

    useEffect(() => {
        let timer: ReturnType<typeof setTimeout> | null = null;
        const off = addWSListener((msg: WSMessage) => {
            if (msg.type !== "vault_changed") {
                return;
            }
            if (timer) {
                clearTimeout(timer);
            }
            timer = setTimeout(() => {
                qc.invalidateQueries({ queryKey: ["vault"] });
            }, 250);
        });
        return () => {
            if (timer) {
                clearTimeout(timer);
            }
            off();
        };
    }, [addWSListener, qc]);

    function onError(err: unknown) {
        setActionError(err instanceof Error ? err.message : "Action failed");
    }

    function clearSelection() {
        setSelected(new Set());
    }

    function toggleSelect(key: string) {
        setSelected(prev => {
            const next = new Set(prev);
            if (next.has(key)) {
                next.delete(key);
            } else {
                next.add(key);
            }
            return next;
        });
    }

    function toggleSelectAll() {
        if (allSelected) {
            clearSelection();
            return;
        }

        const next = new Set<string>();
        for (let i = 0; i < folders.length; i++) {
            next.add(folderKey(folders[i].id));
        }
        for (let i = 0; i < files.length; i++) {
            next.add(fileKey(files[i].id));
        }
        setSelected(next);
    }

    function navigateToFolder(id: string | null) {
        setActionError(null);
        clearSelection();
        if (id === null) {
            setSearchParams({});
            return;
        }

        setSearchParams({ folder: id });
    }

    function canManageItem(ownerId: string): boolean {
        if (staff) {
            return true;
        }

        return myId === ownerId;
    }

    function setTaskStatus(taskId: number, patch: Partial<UploadTask>) {
        setUploads(prev => prev.map(task => (task.id === taskId ? { ...task, ...patch } : task)));
    }

    function uploadFiles(fileList: FileList | null) {
        if (!fileList || fileList.length === 0) {
            return;
        }

        setActionError(null);

        const tasks: UploadTask[] = [];
        for (let i = 0; i < fileList.length; i++) {
            tasks.push({
                id: uploadIdRef.current++,
                name: fileList[i].name,
                file: fileList[i],
                progress: 0,
                status: "pending",
            });
        }
        setUploads(prev => [...prev, ...tasks]);

        const url = vaultUploadUrl();
        const targetFolder = folderId;
        const queue = [...tasks];

        const worker = async () => {
            while (queue.length > 0) {
                const task = queue.shift();
                if (!task) {
                    break;
                }

                setTaskStatus(task.id, { status: "uploading" });
                try {
                    await uploadOne(url, task.file, targetFolder, pct => setTaskStatus(task.id, { progress: pct }));
                    setTaskStatus(task.id, { status: "done", progress: 100 });
                } catch {
                    setTaskStatus(task.id, { status: "error" });
                }
            }
        };

        const run = async () => {
            const workers: Promise<void>[] = [];
            const poolSize = Math.min(UPLOAD_CONCURRENCY, tasks.length);
            for (let i = 0; i < poolSize; i++) {
                workers.push(worker());
            }
            await Promise.all(workers);
            await qc.invalidateQueries({ queryKey: ["vault"] });
        };

        run().catch(() => setActionError("Some uploads failed"));
    }

    function onFilesSelected(event: ChangeEvent<HTMLInputElement>) {
        uploadFiles(event.target.files);
        event.target.value = "";
    }

    function onDrop(event: DragEvent<HTMLDivElement>) {
        event.preventDefault();
        setDragActive(false);
        uploadFiles(event.dataTransfer.files);
    }

    function selectedFolders(): VaultFolder[] {
        return folders.filter(folder => selected.has(folderKey(folder.id)));
    }

    function selectedFiles(): VaultFile[] {
        return files.filter(file => selected.has(fileKey(file.id)));
    }

    function bulkSetLocked(locked: boolean) {
        setActionError(null);
        const fs = selectedFolders();
        for (let i = 0; i < fs.length; i++) {
            lockFolder.mutate({ id: fs[i].id, locked }, { onError });
        }
        const items = selectedFiles();
        for (let i = 0; i < items.length; i++) {
            lockFile.mutate({ id: items[i].id, locked }, { onError });
        }
        clearSelection();
    }

    function confirmBulkDelete() {
        setActionError(null);
        const fs = selectedFolders();
        for (let i = 0; i < fs.length; i++) {
            deleteFolder.mutate(fs[i].id, { onError });
        }
        const items = selectedFiles();
        for (let i = 0; i < items.length; i++) {
            deleteFile.mutate(items[i].id, { onError });
        }
        clearSelection();
        setBulkDeleteOpen(false);
    }

    function submitCreate() {
        const name = createName.trim();
        if (name === "") {
            return;
        }

        createFolder.mutate(
            { name, parentId: folderId },
            {
                onError,
                onSuccess: () => {
                    setCreateOpen(false);
                    setCreateName("");
                },
            },
        );
    }

    function submitRename() {
        if (!renameTarget) {
            return;
        }

        const name = renameValue.trim();
        if (name === "") {
            return;
        }

        const mutation = renameTarget.kind === "folder" ? renameFolder : renameFile;
        mutation.mutate(
            { id: renameTarget.id, name },
            {
                onError,
                onSuccess: () => setRenameTarget(null),
            },
        );
    }

    function confirmDelete() {
        if (!deleteTarget) {
            return;
        }

        const mutation = deleteTarget.kind === "folder" ? deleteFolder : deleteFile;
        mutation.mutate(deleteTarget.id, {
            onError,
            onSuccess: () => setDeleteTarget(null),
        });
    }

    function openFolderMenu(event: MouseEvent, folder: VaultFolder) {
        const items: ContextMenuItem[] = [
            { id: "open", label: "Open", icon: "▸", onClick: () => navigateToFolder(folder.id) },
        ];

        if (canManageLocks) {
            items.push({
                id: "lock",
                label: folder.locked ? "Unlock folder" : "Lock folder",
                icon: folder.locked ? "🔓" : "🔒",
                onClick: () => lockFolder.mutate({ id: folder.id, locked: !folder.locked }, { onError }),
            });
        }

        if (canManageItem(folder.createdBy)) {
            items.push({
                id: "rename",
                label: "Rename",
                icon: "✎",
                onClick: () => {
                    setRenameTarget({ kind: "folder", id: folder.id, name: folder.name });
                    setRenameValue(folder.name);
                },
            });
            items.push({
                id: "delete",
                label: "Delete",
                icon: "✕",
                variant: "danger",
                separator: true,
                onClick: () => setDeleteTarget({ kind: "folder", id: folder.id, name: folder.name }),
            });
        }

        openMenu(event, items);
    }

    function openFileMenu(event: MouseEvent, file: VaultFile) {
        const items: ContextMenuItem[] = [
            {
                id: "download",
                label: "Download",
                icon: "↓",
                onClick: () => {
                    window.location.href = vaultDownloadUrl(file.id);
                },
            },
        ];

        if (canManageLocks) {
            items.push({
                id: "lock",
                label: file.locked ? "Unlock file" : "Lock file",
                icon: file.locked ? "🔓" : "🔒",
                onClick: () => lockFile.mutate({ id: file.id, locked: !file.locked }, { onError }),
            });
        }

        if (canManageItem(file.uploadedBy)) {
            items.push({
                id: "rename",
                label: "Rename",
                icon: "✎",
                onClick: () => {
                    setRenameTarget({ kind: "file", id: file.id, name: file.name });
                    setRenameValue(file.name);
                },
            });
            items.push({
                id: "delete",
                label: "Delete",
                icon: "✕",
                variant: "danger",
                separator: true,
                onClick: () => setDeleteTarget({ kind: "file", id: file.id, name: file.name }),
            });
        }

        openMenu(event, items);
    }

    return (
        <div className={styles.shell}>
            {!isMobile && <ChannelRail />}
            <div
                className={styles.main}
                onDragEnter={event => {
                    if (event.dataTransfer.types.includes("Files")) {
                        event.preventDefault();
                        setDragActive(true);
                    }
                }}
                onDragOver={event => {
                    if (dragActive) {
                        event.preventDefault();
                    }
                }}
            >
                <div className={styles.page}>
                    <div className={styles.head}>
                        <h1 className={styles.title}>Data Vault</h1>
                        {canManageLocks && <span className={styles.gmBadge}>GM access</span>}
                    </div>
                    <p className={styles.subtitle}>
                        Shared file storage. Click a folder to open it, a file to download. Locked folders are visible
                        to GMs and staff only.
                    </p>

                    <nav className={styles.breadcrumbs} aria-label="Breadcrumb">
                        <button
                            type="button"
                            className={`${styles.crumb}${crumbs.length === 0 ? ` ${styles.crumbCurrent}` : ""}`}
                            onClick={() => navigateToFolder(null)}
                        >
                            <span className={styles.crumbIcon} aria-hidden="true">
                                {"📁"}
                            </span>
                            root
                        </button>
                        {crumbs.map((crumb, index) => (
                            <span key={crumb.id} className={styles.crumbWrap}>
                                <span className={styles.crumbSep} aria-hidden="true">
                                    {"›"}
                                </span>
                                <button
                                    type="button"
                                    className={`${styles.crumb}${index === crumbs.length - 1 ? ` ${styles.crumbCurrent}` : ""}`}
                                    onClick={() => navigateToFolder(crumb.id)}
                                >
                                    {crumb.name}
                                </button>
                            </span>
                        ))}
                        {itemCount > 0 && (
                            <span className={styles.count}>
                                {itemCount} item{itemCount === 1 ? "" : "s"}
                            </span>
                        )}
                    </nav>

                    <div className={styles.toolbar}>
                        <Button variant="primary" size="small" onClick={() => setCreateOpen(true)}>
                            + New Folder
                        </Button>
                        <Button variant="ghost" size="small" onClick={() => fileInputRef.current?.click()}>
                            ↑ Upload
                        </Button>
                        <input
                            ref={fileInputRef}
                            type="file"
                            multiple
                            className={styles.hiddenInput}
                            onChange={onFilesSelected}
                        />
                    </div>

                    {uploads.length > 0 && (
                        <div className={styles.uploadPanel}>
                            <div className={styles.uploadHead}>
                                <span>
                                    {uploadActive
                                        ? `Uploading ${uploadDone}/${uploads.length}…`
                                        : `Uploaded ${uploadDone}/${uploads.length}${uploadFailed > 0 ? ` · ${uploadFailed} failed` : ""}`}
                                </span>
                                {!uploadActive && (
                                    <button type="button" className={styles.uploadClear} onClick={() => setUploads([])}>
                                        Clear
                                    </button>
                                )}
                            </div>
                            <div className={styles.uploadList}>
                                {uploads.map(task => (
                                    <div key={task.id} className={styles.uploadRow}>
                                        <span className={styles.uploadName} title={task.name}>
                                            {task.name}
                                        </span>
                                        <div className={styles.uploadBarWrap}>
                                            <div
                                                className={styles.uploadBar}
                                                data-status={task.status}
                                                style={{ width: `${task.status === "done" ? 100 : task.progress}%` }}
                                            />
                                        </div>
                                        <span className={styles.uploadStat} data-status={task.status}>
                                            {uploadStatLabel(task)}
                                        </span>
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}

                    {selected.size > 0 && (
                        <div className={styles.selectionBar}>
                            <span className={styles.selectionCount}>{selected.size} selected</span>
                            {canManageLocks && (
                                <Button variant="ghost" size="small" onClick={() => bulkSetLocked(true)}>
                                    🔒 Lock
                                </Button>
                            )}
                            {canManageLocks && (
                                <Button variant="ghost" size="small" onClick={() => bulkSetLocked(false)}>
                                    🔓 Unlock
                                </Button>
                            )}
                            <Button variant="danger" size="small" onClick={() => setBulkDeleteOpen(true)}>
                                Delete
                            </Button>
                            <Button variant="ghost" size="small" onClick={clearSelection}>
                                Clear
                            </Button>
                        </div>
                    )}

                    {actionError && (
                        <div className={styles.error} role="alert">
                            {actionError}
                        </div>
                    )}

                    <div className={styles.dropZone}>
                        {browse.loading && <div className={styles.placeholder}>Loading vault…</div>}
                        {isEmpty && (
                            <div className={styles.placeholder}>This folder is empty. Drop files here or upload.</div>
                        )}

                        {!browse.loading && itemCount > 0 && (
                            <div className={styles.list}>
                                <div className={styles.listHead}>
                                    <input
                                        type="checkbox"
                                        className={styles.rowCheck}
                                        checked={allSelected}
                                        onChange={toggleSelectAll}
                                        aria-label="Select all"
                                    />
                                    <span className={styles.headName}>Name</span>
                                    <span className={styles.headMeta}>Size</span>
                                    <span className={styles.headActions} aria-hidden="true" />
                                </div>

                                {folders.map(folder => {
                                    const key = folderKey(folder.id);
                                    return (
                                        <div
                                            key={folder.id}
                                            className={`${styles.row} ${styles.folderRow}${selected.has(key) ? ` ${styles.rowSelected}` : ""}`}
                                            onContextMenu={event => openFolderMenu(event, folder)}
                                        >
                                            <input
                                                type="checkbox"
                                                className={styles.rowCheck}
                                                checked={selected.has(key)}
                                                onChange={() => toggleSelect(key)}
                                                aria-label={`Select ${folder.name}`}
                                            />
                                            <button
                                                type="button"
                                                className={styles.rowMain}
                                                onClick={() => navigateToFolder(folder.id)}
                                                title={folder.name}
                                            >
                                                <span className={styles.folderIcon} aria-hidden="true">
                                                    {"📁"}
                                                </span>
                                                <span className={styles.rowName}>{folder.name}</span>
                                                {folder.locked && (
                                                    <span className={styles.lockBadge} title="Locked - GM only">
                                                        {"🔒"}
                                                    </span>
                                                )}
                                                <span className={styles.rowMeta}>folder</span>
                                            </button>
                                            <button
                                                type="button"
                                                className={styles.rowActions}
                                                onClick={event => openFolderMenu(event, folder)}
                                                aria-label="Folder actions"
                                            >
                                                {"⋯"}
                                            </button>
                                        </div>
                                    );
                                })}

                                {files.map(file => {
                                    const key = fileKey(file.id);
                                    return (
                                        <div
                                            key={file.id}
                                            className={`${styles.row}${selected.has(key) ? ` ${styles.rowSelected}` : ""}`}
                                            onContextMenu={event => openFileMenu(event, file)}
                                        >
                                            <input
                                                type="checkbox"
                                                className={styles.rowCheck}
                                                checked={selected.has(key)}
                                                onChange={() => toggleSelect(key)}
                                                aria-label={`Select ${file.name}`}
                                            />
                                            <a
                                                className={styles.rowMain}
                                                href={vaultDownloadUrl(file.id)}
                                                download
                                                title={file.name}
                                            >
                                                <span className={styles.fileIcon} aria-hidden="true">
                                                    {"📄"}
                                                </span>
                                                <span className={styles.rowName}>{file.name}</span>
                                                {file.locked && (
                                                    <span className={styles.lockBadge} title="Locked - GM only">
                                                        {"🔒"}
                                                    </span>
                                                )}
                                                <span className={styles.rowMeta}>{formatSize(file.size)}</span>
                                            </a>
                                            <button
                                                type="button"
                                                className={styles.rowActions}
                                                onClick={event => openFileMenu(event, file)}
                                                aria-label="File actions"
                                            >
                                                {"⋯"}
                                            </button>
                                        </div>
                                    );
                                })}
                            </div>
                        )}
                    </div>

                    <Modal
                        isOpen={createOpen}
                        onClose={() => {
                            setCreateOpen(false);
                            setCreateName("");
                        }}
                        title="New Folder"
                    >
                        <div className={styles.modalBody}>
                            <input
                                className={styles.modalInput}
                                autoFocus
                                placeholder="Folder name"
                                value={createName}
                                onChange={event => setCreateName(event.target.value)}
                                onKeyDown={event => {
                                    if (event.key === "Enter") {
                                        submitCreate();
                                    }
                                }}
                            />
                            <div className={styles.modalActions}>
                                <Button variant="ghost" size="small" onClick={() => setCreateOpen(false)}>
                                    Cancel
                                </Button>
                                <Button variant="primary" size="small" onClick={submitCreate}>
                                    Create
                                </Button>
                            </div>
                        </div>
                    </Modal>

                    <Modal isOpen={renameTarget !== null} onClose={() => setRenameTarget(null)} title="Rename">
                        <div className={styles.modalBody}>
                            <input
                                className={styles.modalInput}
                                autoFocus
                                value={renameValue}
                                onChange={event => setRenameValue(event.target.value)}
                                onKeyDown={event => {
                                    if (event.key === "Enter") {
                                        submitRename();
                                    }
                                }}
                            />
                            <div className={styles.modalActions}>
                                <Button variant="ghost" size="small" onClick={() => setRenameTarget(null)}>
                                    Cancel
                                </Button>
                                <Button variant="primary" size="small" onClick={submitRename}>
                                    Rename
                                </Button>
                            </div>
                        </div>
                    </Modal>

                    <Modal isOpen={deleteTarget !== null} onClose={() => setDeleteTarget(null)} title="Delete">
                        <div className={styles.modalBody}>
                            <p className={styles.confirmText}>
                                Delete <strong>{deleteTarget?.name}</strong>?
                                {deleteTarget?.kind === "folder" && " Everything inside it is removed too."} This
                                can&apos;t be undone.
                            </p>
                            <div className={styles.modalActions}>
                                <Button variant="ghost" size="small" onClick={() => setDeleteTarget(null)}>
                                    Cancel
                                </Button>
                                <Button variant="danger" size="small" onClick={confirmDelete}>
                                    Delete
                                </Button>
                            </div>
                        </div>
                    </Modal>

                    <Modal isOpen={bulkDeleteOpen} onClose={() => setBulkDeleteOpen(false)} title="Delete selected">
                        <div className={styles.modalBody}>
                            <p className={styles.confirmText}>
                                Delete <strong>{selected.size}</strong> selected item{selected.size === 1 ? "" : "s"}?
                                Any selected folders remove everything inside them too. This can&apos;t be undone.
                            </p>
                            <div className={styles.modalActions}>
                                <Button variant="ghost" size="small" onClick={() => setBulkDeleteOpen(false)}>
                                    Cancel
                                </Button>
                                <Button variant="danger" size="small" onClick={confirmBulkDelete}>
                                    Delete {selected.size}
                                </Button>
                            </div>
                        </div>
                    </Modal>

                    <ContextMenu state={menuState} onClose={closeMenu} />
                </div>

                {dragActive && (
                    <div
                        className={styles.dropOverlay}
                        onDragOver={event => event.preventDefault()}
                        onDragLeave={() => setDragActive(false)}
                        onDrop={onDrop}
                    >
                        <span className={styles.dropOverlayText}>Drop files to upload</span>
                    </div>
                )}
            </div>
        </div>
    );
}
