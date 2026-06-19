import type { ChatRoomMember } from "../types/api";
import { ROLE_GROUPS } from "./permissions";

export interface MemberGroup {
    label: string;
    members: ChatRoomMember[];
}

export function buildMemberGroups(
    members: ChatRoomMember[],
    memberOnlineWeight: (id: string) => number,
    voiceIdSet: Set<string>,
): MemberGroup[] {
    const memberRankWeight = (m: ChatRoomMember) => {
        if (m.user.role === "super_admin") {
            return 0;
        }

        if (m.role === "host") {
            return 1;
        }

        const idx = ROLE_GROUPS.findIndex(g => g.role === m.user.role);
        if (idx >= 0) {
            return idx + 1;
        }

        return ROLE_GROUPS.length + 1;
    };
    const memberSortName = (m: ChatRoomMember) => {
        const nickname = m.nickname?.trim();
        if (nickname) {
            return nickname.toLowerCase();
        }
        const displayName = m.user.display_name?.trim();
        if (displayName) {
            return displayName.toLowerCase();
        }
        return m.user.username.toLowerCase();
    };
    const memberRankLabel = (m: ChatRoomMember) => {
        const group = ROLE_GROUPS.find(g => g.role === m.user.role);
        if (m.user.role === "super_admin" && group) {
            return group.label;
        }

        if (m.role === "host") {
            return "Host";
        }

        if (group) {
            return group.label;
        }

        return "Members";
    };
    const sortedMembers = [...members].sort((a, b) => {
        const rank = memberRankWeight(a) - memberRankWeight(b);
        if (rank !== 0) {
            return rank;
        }

        const online = memberOnlineWeight(a.user.id) - memberOnlineWeight(b.user.id);
        if (online !== 0) {
            return online;
        }

        return memberSortName(a).localeCompare(memberSortName(b));
    });
    const memberGroups: MemberGroup[] = [];
    for (const m of sortedMembers) {
        const label = memberRankLabel(m);
        const last = memberGroups[memberGroups.length - 1];
        if (last && last.label === label) {
            last.members.push(m);
        } else {
            memberGroups.push({ label, members: [m] });
        }
    }

    const inVoiceMembers = sortedMembers.filter(m => voiceIdSet.has(m.user.id));
    if (inVoiceMembers.length > 0) {
        memberGroups.unshift({ label: "In Voice", members: inVoiceMembers });
    }

    return memberGroups;
}
