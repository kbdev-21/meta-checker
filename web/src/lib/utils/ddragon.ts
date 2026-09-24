// Backend chưa trả version ddragon riêng; lấy từ imgUrl của champion (.../cdn/<version>/img/...).
export function ddragonVersionOf(imgUrl: string): string | null {
	return imgUrl.match(/\/cdn\/([^/]+)\//)?.[1] ?? null;
}

export function profileIconUrl(version: string, iconId: number): string {
	return `https://ddragon.leagueoflegends.com/cdn/${version}/img/profileicon/${iconId}.png`;
}
