import { useNavigate, useParams } from "react-router";
import { useAuth } from "../../hooks/useAuth";
import { useProfile } from "../../api/queries/profile";
import { usePageTitle } from "../../hooks/usePageTitle";
import { useBlock } from "../../hooks/useBlock";
import { parseServerDate } from "../../utils/time";
import { Button } from "../../components/Button/Button";
import { RolePill } from "../../components/RolePill/RolePill";
import { RoleStyledName } from "../../components/RoleStyledName/RoleStyledName";
import styles from "./ProfilePage.module.css";

const SOCIAL_LABELS: Record<string, string> = {
    social_twitter: "Twitter / X",
    social_discord: "Discord",
    social_waifulist: "WaifuList",
    social_tumblr: "Tumblr",
    social_github: "GitHub",
};

function formatDate(iso: string): string {
    const d = parseServerDate(iso);
    if (!d) {
        return "";
    }
    return d.toLocaleDateString(undefined, { year: "numeric", month: "long", day: "numeric" });
}

function formatDOBWithAge(value: string): string {
    const parts = value.split("-");
    if (parts.length !== 3) {
        return value;
    }

    const year = Number(parts[0]);
    const month = Number(parts[1]);
    const day = Number(parts[2]);

    if (!Number.isInteger(year) || !Number.isInteger(month) || !Number.isInteger(day)) {
        return value;
    }

    const parsed = new Date(Date.UTC(year, month - 1, day));
    if (Number.isNaN(parsed.getTime())) {
        return value;
    }

    const now = new Date();
    let age = now.getUTCFullYear() - year;
    if (now.getUTCMonth() + 1 < month || (now.getUTCMonth() + 1 === month && now.getUTCDate() < day)) {
        age -= 1;
    }

    const ageLabel = age === 1 ? "year old" : "years old";
    const formatted = parsed.toLocaleDateString(undefined, {
        year: "numeric",
        month: "long",
        day: "2-digit",
        timeZone: "UTC",
    });

    if (age < 0) {
        return formatted;
    }

    return `${formatted} (${age} ${ageLabel})`;
}

function socialUrl(key: string, value: string): string {
    if (value.startsWith("http://") || value.startsWith("https://")) {
        return value;
    }
    switch (key) {
        case "social_twitter":
            return `https://x.com/${value}`;
        case "social_github":
            return `https://github.com/${value}`;
        case "social_tumblr":
            return `https://${value}.tumblr.com`;
        case "social_waifulist":
            return value.includes("/") ? `https://${value}` : value;
        default:
            return value;
    }
}

export function ProfilePage() {
    const { username } = useParams<{ username: string }>();
    const navigate = useNavigate();
    const { user: currentUser } = useAuth();
    const { profile, loading } = useProfile(username ?? "");
    usePageTitle(profile?.display_name ?? "Profile");
    const blockHook = useBlock(profile?.id ?? "");

    if (loading) {
        return <div className="loading">Pinging the node...</div>;
    }

    if (!profile) {
        return (
            <div className="empty-state">
                Runner not found on this node.
                <br />
                <Button variant="secondary" onClick={() => navigate("/")}>
                    Return
                </Button>
            </div>
        );
    }

    const socialEntries = Object.entries(SOCIAL_LABELS)
        .map(([key, label]) => ({
            key,
            label,
            value: profile[key as keyof typeof profile] as string,
        }))
        .filter(entry => entry.value);

    if (profile.website) {
        socialEntries.push({ key: "website", label: "Website", value: profile.website });
    }

    if (profile.email) {
        socialEntries.push({ key: "email", label: "Email", value: profile.email });
    }

    const showGender = profile.gender && profile.gender !== "Prefer not to say";
    const isBanned = profile.banned === true;

    return (
        <div className={`${styles.page} ${isBanned ? styles.bannedProfile : ""}`}>
            {isBanned && (
                <div className={styles.banBanner}>
                    <span className={styles.banBannerTitle}>SIN revoked // runner blacklisted</span>
                    {profile.ban_reason && <span className={styles.banBannerReason}>Reason: {profile.ban_reason}</span>}
                </div>
            )}
            <div className={styles.banner}>
                {profile.banner_url ? (
                    <img
                        src={profile.banner_url}
                        alt=""
                        className={styles.bannerImage}
                        style={{ objectPosition: `center ${profile.banner_position ?? 50}%` }}
                    />
                ) : (
                    <div className={styles.bannerGradient} />
                )}
            </div>

            <div className={styles.headerSection}>
                <div className={styles.avatarContainer}>
                    {profile.avatar_url ? (
                        <img src={profile.avatar_url} alt={profile.display_name} className={styles.avatar} />
                    ) : (
                        <div className={styles.avatarPlaceholder}>{profile.display_name.charAt(0).toUpperCase()}</div>
                    )}
                    {profile.online && <span className={styles.onlineDot} />}
                </div>
                <div className={styles.info}>
                    <h1 className={styles.displayName}>
                        <RoleStyledName name={profile.display_name} role={profile.role} />
                        <RolePill role={profile.role ?? ""} userId={profile.id} />
                    </h1>
                    <span className={styles.username}>@{profile.username}</span>
                    {currentUser && currentUser.id !== profile.id && (
                        <div className={styles.followRow}>
                            {blockHook.status && !profile.role && (
                                <Button variant="ghost" size="small" onClick={blockHook.toggleBlock}>
                                    {blockHook.status.blocking ? "Unblock" : "Block"}
                                </Button>
                            )}
                        </div>
                    )}
                    {blockHook.status?.blocked_by && (
                        <div className={styles.blockedBanner}>This runner has blocked you.</div>
                    )}
                    <div className={styles.metaRow}>
                        {showGender && <span className={styles.metaItem}>{profile.gender}</span>}
                        {profile.pronoun_subject && profile.pronoun_possessive && (
                            <span className={styles.metaItem}>
                                {profile.pronoun_subject}/{profile.pronoun_possessive}
                            </span>
                        )}
                        {profile.dob && <span className={styles.metaItem}>Born {formatDOBWithAge(profile.dob)}</span>}
                        <span className={styles.metaItem}>Joined {formatDate(profile.created_at)}</span>
                    </div>
                </div>
            </div>

            <div className={styles.bio}>{profile.bio || "This runner has not written a bio yet."}</div>

            {socialEntries.length > 0 && (
                <div className={styles.socialRow}>
                    {socialEntries.map(entry => (
                        <span key={entry.key} className={styles.socialChip}>
                            <span className={styles.socialChipLabel}>{entry.label}</span>
                            {entry.key === "social_discord" ? (
                                <span className={styles.socialChipValue}>{entry.value}</span>
                            ) : (
                                <a
                                    className={styles.socialChipValue}
                                    href={
                                        entry.key === "website"
                                            ? entry.value.startsWith("http")
                                                ? entry.value
                                                : `https://${entry.value}`
                                            : entry.key === "email"
                                              ? `mailto:${entry.value}`
                                              : socialUrl(entry.key, entry.value)
                                    }
                                    target="_blank"
                                    rel="noopener noreferrer"
                                >
                                    {entry.value}
                                </a>
                            )}
                        </span>
                    ))}
                </div>
            )}

            {profile.favourite_character && (
                <div className={styles.favourite}>
                    <span className={styles.favouriteLabel}>Favourite Character</span>
                    <span className={styles.favouriteValue}>{profile.favourite_character}</span>
                </div>
            )}
        </div>
    );
}
