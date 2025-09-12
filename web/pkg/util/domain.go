package util

import (
	"net"
	"regexp"
	"strings"
)

// ValidateDomain 验证域名格式是否合法
func ValidateDomain(domain string) bool {
	if domain == "" {
		return false
	}

	// 基本长度检查
	if len(domain) > 253 {
		return false
	}

	// 域名正则表达式
	domainRegex := `^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+(([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)|(\*))$`
	matched, _ := regexp.MatchString(domainRegex, domain)
	if !matched {
		return false
	}

	// 检查每个标签长度
	labels := strings.Split(domain, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		// 标签不能以连字符开始或结束
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return false
		}
	}

	return true
}

// ValidateDomainAdvanced 高级域名校验（包含DNS解析检查）
func ValidateDomainAdvanced(domain string) (bool, string) {
	// 基础格式校验
	if !ValidateDomain(domain) {
		return false, "域名格式不正确"
	}

	// 检查是否为IP地址
	if net.ParseIP(domain) != nil {
		return false, "不能使用IP地址，请使用域名"
	}

	// 检查顶级域名
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return false, "域名必须包含顶级域名"
	}

	// 黑名单检查
	blacklistDomains := []string{
		"localhost",
		"127.0.0.1",
		"0.0.0.0",
		"example.com",
		"test.com",
		"invalid",
	}

	for _, blackDomain := range blacklistDomains {
		if strings.EqualFold(domain, blackDomain) {
			return false, "该域名不允许添加"
		}
	}

	// 可选：DNS解析检查（可能会影响性能）
	// _, err := net.LookupHost(domain)
	// if err != nil {
	//     return false, "域名无法解析，请检查域名是否存在"
	// }

	return true, ""
}

// SanitizeDomain 清理和标准化域名
func SanitizeDomain(domain string) string {
	// 转换为小写
	domain = strings.ToLower(domain)

	// 移除前后空格
	domain = strings.TrimSpace(domain)

	// 移除协议前缀
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "www.")

	// 移除路径和查询参数
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}
	if idx := strings.Index(domain, "?"); idx != -1 {
		domain = domain[:idx]
	}

	return domain
}
