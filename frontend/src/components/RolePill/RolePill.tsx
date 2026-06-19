import { useEffect, useRef, useState } from "react";
import { useSiteInfo } from "../../hooks/useSiteInfo";
import styles from "./RolePill.module.css";

interface RolePillProps {
    role: string;
    userId?: string;
    compactOnMobile?: boolean;
    compact?: boolean;
}

const roleConfig: Record<string, { label: string; className: string; tooltip: string }> = {
    super_admin: { label: "Sysop", className: "superAdmin", tooltip: "Site owner - super administrator" },
    admin: { label: "Admin", className: "admin", tooltip: "Administrator" },
    moderator: { label: "Moderator", className: "moderator", tooltip: "Moderator" },
};

function hexToRgba(hex: string, alpha: number): string {
    const r = parseInt(hex.slice(1, 3), 16);
    const g = parseInt(hex.slice(3, 5), 16);
    const b = parseInt(hex.slice(5, 7), 16);
    return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}

function darkenForText(hex: string): string {
    const r = parseInt(hex.slice(1, 3), 16) / 255;
    const g = parseInt(hex.slice(3, 5), 16) / 255;
    const b = parseInt(hex.slice(5, 7), 16) / 255;
    const max = Math.max(r, g, b);
    const min = Math.min(r, g, b);
    const l = (max + min) / 2;
    if (l <= 0.42) {
        return hex;
    }
    const scale = 0.42 / l;
    const nr = Math.round(r * scale * 255);
    const ng = Math.round(g * scale * 255);
    const nb = Math.round(b * scale * 255);
    const toHex = (v: number) => v.toString(16).padStart(2, "0");
    return `#${toHex(nr)}${toHex(ng)}${toHex(nb)}`;
}

export function RolePill({ role, userId, compactOnMobile, compact }: RolePillProps) {
    const siteInfo = useSiteInfo();
    const config = roleConfig[role];
    const [expanded, setExpanded] = useState(false);
    const collapseTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    useEffect(() => {
        return () => {
            if (collapseTimerRef.current) {
                clearTimeout(collapseTimerRef.current);
            }
        };
    }, []);

    const userVanityRoleIds = (userId && siteInfo.vanity_role_assignments?.[userId]) ?? [];
    const allVanityRoles = siteInfo.vanity_roles ?? [];
    const vanityRoles = [];
    for (const vr of allVanityRoles) {
        if (userVanityRoleIds.includes(vr.id)) {
            vanityRoles.push(vr);
        }
    }
    vanityRoles.sort((a, b) => a.sort_order - b.sort_order);

    const collapsible = compactOnMobile === true || compact === true;
    const isCompacted = collapsible && !expanded;
    const compactClass = isCompacted ? (compact ? ` ${styles.alwaysCompact}` : ` ${styles.compactMobile}`) : "";

    const onPillClick = collapsible
        ? (e: React.MouseEvent) => {
              e.stopPropagation();
              e.preventDefault();
              if (collapseTimerRef.current) {
                  clearTimeout(collapseTimerRef.current);
                  collapseTimerRef.current = null;
              }
              setExpanded(prev => {
                  const next = !prev;
                  if (next) {
                      collapseTimerRef.current = setTimeout(() => setExpanded(false), 5000);
                  }
                  return next;
              });
          }
        : undefined;

    if (!config && vanityRoles.length === 0) {
        return null;
    }

    const groupClass = `${styles.group}${compactOnMobile ? ` ${styles.groupCompactMobile}` : ""}${
        compact ? ` ${styles.groupAlwaysCompact}` : ""
    }${isCompacted ? "" : ` ${styles.groupExpanded}`}`;

    return (
        <span className={groupClass} aria-label="User roles">
            {config && (
                <span
                    className={`${styles.pill} ${styles[config.className]}${compactClass}`}
                    title={config.tooltip}
                    onClick={onPillClick}
                >
                    {config.label}
                </span>
            )}
            {vanityRoles.map(vr => {
                const tooltip = vr.label;
                return (
                    <span
                        key={vr.id}
                        className={`${styles.pill}${compactClass}`}
                        title={tooltip}
                        style={{
                            backgroundColor: hexToRgba(vr.color, 0.18),
                            color: darkenForText(vr.color),
                            border: `1px solid ${hexToRgba(vr.color, 0.55)}`,
                            ["--dot-color" as string]: vr.color,
                        }}
                        onClick={onPillClick}
                    >
                        {vr.label}
                    </span>
                );
            })}
        </span>
    );
}
