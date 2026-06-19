import React, { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router";
import { useAuth } from "../../hooks/useAuth";
import { useSettingsForm } from "../../hooks/useSettingsForm";
import { usePageTitle } from "../../hooks/usePageTitle";
import { Button } from "../../components/Button/Button";
import { Input } from "../../components/Input/Input";
import { TextArea } from "../../components/TextArea/TextArea";
import { Select } from "../../components/Select/Select";
import { ToggleSwitch } from "../../components/ToggleSwitch/ToggleSwitch";
import { BlockedUsersSection } from "./BlockedUsersSection";
import { ChangePasswordSection } from "./ChangePasswordSection";
import { DangerZoneSection } from "./DangerZoneSection";
import styles from "./SettingsPage.module.css";

function BannerSection({ form }: { form: ReturnType<typeof useSettingsForm> }) {
    const containerRef = useRef<HTMLDivElement>(null);
    const [dragging, setDragging] = useState(false);
    const dragStartY = useRef(0);
    const dragStartPos = useRef(0);

    const handlePointerDown = useCallback(
        (e: React.MouseEvent | React.TouchEvent) => {
            if (!form.bannerUrl) {
                return;
            }
            dragStartY.current = "touches" in e ? e.touches[0].clientY : e.clientY;
            dragStartPos.current = form.bannerPosition;
            setDragging(true);
        },
        [form.bannerUrl, form.bannerPosition],
    );

    const handlePointerMove = useCallback(
        (e: MouseEvent | TouchEvent) => {
            if (!dragging || !containerRef.current) {
                return;
            }
            const clientY = "touches" in e ? e.touches[0].clientY : e.clientY;
            const containerHeight = containerRef.current.getBoundingClientRect().height;
            const deltaPercent = ((clientY - dragStartY.current) / containerHeight) * 100;
            const newPos = Math.min(100, Math.max(0, dragStartPos.current - deltaPercent));
            form.setBannerPosition(newPos);
        },
        [dragging, form],
    );

    const handlePointerUp = useCallback(() => {
        setDragging(false);
    }, []);

    useEffect(() => {
        if (dragging) {
            document.addEventListener("mousemove", handlePointerMove);
            document.addEventListener("mouseup", handlePointerUp);
            document.addEventListener("touchmove", handlePointerMove);
            document.addEventListener("touchend", handlePointerUp);
        }
        return () => {
            document.removeEventListener("mousemove", handlePointerMove);
            document.removeEventListener("mouseup", handlePointerUp);
            document.removeEventListener("touchmove", handlePointerMove);
            document.removeEventListener("touchend", handlePointerUp);
        };
    }, [dragging, handlePointerMove, handlePointerUp]);

    return (
        <div className={styles.section}>
            <h3 className={styles.sectionTitle}>banner</h3>
            <div className={styles.bannerSection}>
                <div
                    ref={containerRef}
                    className={styles.bannerPreview}
                    style={{ cursor: form.bannerUrl ? "grab" : undefined, userSelect: dragging ? "none" : undefined }}
                    onMouseDown={handlePointerDown}
                    onTouchStart={handlePointerDown}
                >
                    {form.bannerUrl ? (
                        <>
                            <img
                                src={form.bannerUrl}
                                alt="Banner"
                                draggable={false}
                                style={{ objectPosition: `center ${form.bannerPosition}%` }}
                            />
                            <div className={styles.bannerDragHint}>Drag to reposition</div>
                        </>
                    ) : (
                        <div className={styles.bannerPlaceholder}>No banner set</div>
                    )}
                </div>
                <label className={styles.uploadBtn}>
                    {form.uploadingBanner ? "Uploading..." : "Upload Banner"}
                    <input
                        type="file"
                        accept="image/*"
                        onChange={form.handleBannerChange}
                        style={{ display: "none" }}
                        disabled={form.uploadingBanner}
                    />
                </label>
            </div>
        </div>
    );
}

export function SettingsPage() {
    usePageTitle("Settings");
    const navigate = useNavigate();
    const { user, loading: authLoading } = useAuth();
    const form = useSettingsForm();

    useEffect(() => {
        if (!authLoading && !user) {
            navigate("/login");
        }
    }, [user, authLoading, navigate]);

    if (!user) {
        return null;
    }

    if (form.profileLoading) {
        return <div className="loading">Decrypting dossier...</div>;
    }

    return (
        <div className={styles.page}>
            <h2 className={styles.heading}>Settings</h2>
            <p className={styles.subheading}>configure your runner dossier and node preferences</p>

            <form onSubmit={form.handleSubmit}>
                <div className={styles.grid}>
                    <div className={styles.section}>
                        <h3 className={styles.sectionTitle}>avatar</h3>
                        <div className={styles.avatarSection}>
                            <div className={styles.avatarPreview}>
                                {form.avatarUrl ? (
                                    <img src={form.avatarUrl} alt="Avatar" />
                                ) : (
                                    <div className={styles.avatarPlaceholder}>
                                        {form.displayName ? form.displayName.charAt(0).toUpperCase() : "?"}
                                    </div>
                                )}
                            </div>
                            <label className={styles.uploadBtn}>
                                {form.uploadingAvatar ? "Uploading..." : "Upload Avatar"}
                                <input
                                    type="file"
                                    accept="image/*"
                                    onChange={form.handleAvatarChange}
                                    style={{ display: "none" }}
                                    disabled={form.uploadingAvatar}
                                />
                            </label>
                        </div>
                    </div>

                    <BannerSection form={form} />

                    <div className={`${styles.section} ${styles.gridFull}`}>
                        <h3 className={styles.sectionTitle}>dossier</h3>
                        <div className={styles.twoCol}>
                            <label className={styles.label}>
                                Display Name
                                <Input
                                    type="text"
                                    fullWidth
                                    value={form.displayName}
                                    onChange={e => form.setDisplayName(e.target.value)}
                                />
                            </label>
                            <label className={styles.label}>
                                Favourite Character
                                <Input
                                    type="text"
                                    fullWidth
                                    maxLength={100}
                                    placeholder="Favourite character"
                                    value={form.favouriteCharacter}
                                    onChange={e => form.setFavouriteCharacter(e.target.value)}
                                />
                            </label>
                            <label className={styles.label}>
                                Date of Birth
                                <Input
                                    type="date"
                                    fullWidth
                                    value={form.dob}
                                    onChange={e => form.setDob(e.target.value)}
                                />
                            </label>
                        </div>
                        <div>
                            <label className={styles.label}>
                                Gender
                                <Select
                                    value={form.gender}
                                    onChange={e => form.handleGenderChange((e.target as HTMLSelectElement).value)}
                                >
                                    {form.genderOptions.map(opt => (
                                        <option key={opt} value={opt}>
                                            {opt}
                                        </option>
                                    ))}
                                </Select>
                            </label>
                            {form.gender === "Custom" && (
                                <label className={styles.label}>
                                    Custom Gender
                                    <Input
                                        type="text"
                                        fullWidth
                                        value={form.customGender}
                                        onChange={e => form.setCustomGender(e.target.value)}
                                        placeholder="Enter your gender"
                                    />
                                </label>
                            )}
                            <div className={styles.pronounRow}>
                                <span className={styles.pronounPreview}>
                                    Pronouns: {form.pronounSubject}/{form.pronounPossessive}
                                </span>
                                <ToggleSwitch
                                    enabled={form.customPronouns}
                                    onChange={form.handleCustomPronounsToggle}
                                    label="Custom pronouns"
                                />
                            </div>
                            {form.customPronouns && (
                                <div className={styles.twoCol}>
                                    <label className={styles.label}>
                                        Subject (e.g. she, he, they)
                                        <Input
                                            type="text"
                                            fullWidth
                                            value={form.pronounSubject}
                                            onChange={e => form.setPronounSubject(e.target.value.slice(0, 10))}
                                            placeholder="they"
                                            maxLength={10}
                                        />
                                    </label>
                                    <label className={styles.label}>
                                        Possessive (e.g. her, his, their)
                                        <Input
                                            type="text"
                                            fullWidth
                                            value={form.pronounPossessive}
                                            onChange={e => form.setPronounPossessive(e.target.value.slice(0, 10))}
                                            placeholder="their"
                                            maxLength={10}
                                        />
                                    </label>
                                </div>
                            )}
                        </div>
                        <ToggleSwitch
                            enabled={form.dmsEnabled}
                            onChange={form.setDmsEnabled}
                            label="Direct Messages (coming soon)"
                            description="Direct messaging isn't available yet — this setting is inactive for now."
                            disabled
                        />
                        <ToggleSwitch
                            enabled={form.dobPublic}
                            onChange={form.setDobPublic}
                            label="Public Date of Birth"
                            description="Show your date of birth on your public profile"
                        />
                        <label className={styles.label}>
                            Bio
                            <TextArea
                                value={form.bio}
                                onChange={e => form.setBio(e.target.value)}
                                rows={3}
                                placeholder="Drop your handle's backstory for the rest of the node..."
                            />
                        </label>
                    </div>

                    <div className={`${styles.section} ${styles.gridFull}`}>
                        <h3 className={styles.sectionTitle}>email</h3>
                        <label className={styles.label}>
                            Email Address
                            <Input
                                type="email"
                                fullWidth
                                value={form.email}
                                onChange={e => form.setEmail(e.target.value)}
                                placeholder="your@email.com"
                            />
                        </label>
                        <ToggleSwitch
                            enabled={form.emailPublic}
                            onChange={form.setEmailPublic}
                            label="Public Email"
                            description="Show your email address on your profile"
                        />
                    </div>

                    <div className={`${styles.section} ${styles.gridFull}`}>
                        <h3 className={styles.sectionTitle}>alerts</h3>
                        <ToggleSwitch
                            enabled={form.emailNotifications}
                            onChange={form.setEmailNotifications}
                            label="Email Notifications"
                            description="Receive email notifications for replies and upvotes on your posts"
                        />
                        <ToggleSwitch
                            enabled={form.playMessageSound}
                            onChange={form.setPlayMessageSound}
                            label="Chat Message Sound"
                            description="Play a sound for new chat messages when the tab is in the background, except in muted rooms"
                        />
                        <ToggleSwitch
                            enabled={form.playNotificationSound}
                            onChange={form.setPlayNotificationSound}
                            label="Notification Sound"
                            description="Play a sound when a new notification arrives"
                        />
                    </div>

                    <div className={`${styles.section} ${styles.gridFull}`}>
                        <h3 className={styles.sectionTitle}>social links</h3>
                        <div className={styles.twoCol}>
                            <label className={styles.label}>
                                Twitter / X
                                <Input
                                    type="text"
                                    fullWidth
                                    value={form.socialTwitter}
                                    onChange={e => form.setSocialTwitter(e.target.value)}
                                    placeholder="username"
                                />
                            </label>
                            <label className={styles.label}>
                                Discord
                                <Input
                                    type="text"
                                    fullWidth
                                    value={form.socialDiscord}
                                    onChange={e => form.setSocialDiscord(e.target.value)}
                                    placeholder="username#0000"
                                />
                            </label>
                            <label className={styles.label}>
                                WaifuList
                                <Input
                                    type="text"
                                    fullWidth
                                    value={form.socialWaifulist}
                                    onChange={e => form.setSocialWaifulist(e.target.value)}
                                    placeholder="https://waifulist.moe/list/..."
                                />
                            </label>
                            <label className={styles.label}>
                                Tumblr
                                <Input
                                    type="text"
                                    fullWidth
                                    value={form.socialTumblr}
                                    onChange={e => form.setSocialTumblr(e.target.value)}
                                    placeholder="username"
                                />
                            </label>
                            <label className={styles.label}>
                                GitHub
                                <Input
                                    type="text"
                                    fullWidth
                                    value={form.socialGithub}
                                    onChange={e => form.setSocialGithub(e.target.value)}
                                    placeholder="username"
                                />
                            </label>
                            <label className={styles.label}>
                                Website
                                <Input
                                    type="text"
                                    fullWidth
                                    value={form.website}
                                    onChange={e => form.setWebsite(e.target.value)}
                                    placeholder="https://example.com"
                                />
                            </label>
                        </div>
                    </div>
                </div>

                <Button variant="primary" type="submit" disabled={form.saving} style={{ width: "100%" }}>
                    {form.saving ? "Saving..." : "Save Changes"}
                </Button>
                {form.error && <div className={styles.error}>{form.error}</div>}
                {form.success && <div className={styles.success}>{form.success}</div>}
            </form>

            <div className={styles.grid} style={{ marginTop: "1.5rem" }}>
                <BlockedUsersSection />
                <ChangePasswordSection />
                <DangerZoneSection />
            </div>
        </div>
    );
}
