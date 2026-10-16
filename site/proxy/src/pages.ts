export const proxyErrorPage = (code: number, reason: string): string => `<!doctype html>
<html>
<head><meta charset="utf-8"><title>${code} ${reason}</title></head>
<body>
<center><h1>${code} ${reason}</h1></center>
<hr><center>tsubame-proxy</center>
</body>
</html>
`;
