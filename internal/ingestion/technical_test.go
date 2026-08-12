package ingestion

import "testing"

func TestLooksTechnicalAcceptsRealRoles(t *testing.T) {
	// Every title here was taken from a live feed.
	roles := []string{
		"Senior React Full stack Developer",
		"Senior Software QA Engineer",
		"Artificial Intelligence Specialist",
		"Tech Lead Full-Stack Rails Engineer",
		"Senior DevOps Engineer",
		"Senior Product Engineer (Fullstack)",
		"Tier III Service Desk Engineer",
		"Backend Developer",
		"Desenvolvedor Full-Stack Pleno",
		"Pessoa Desenvolvedora Front-end",
		"Engenheiro de Dados",
		"C# Developer",
		"C++ Developer",
		".NET Architect",
		"UX Designer",
		"Site Reliability Engineer",
	}

	for _, role := range roles {
		if !looksTechnical(role) {
			t.Errorf("looksTechnical(%q) = false, want true", role)
		}
	}
}

func TestLooksTechnicalRejectsTheNoise(t *testing.T) {
	// So was every title here, from the same feed.
	roles := []string{
		"Merchandising Execution Associate MARLBOROUGH",
		"factory labourer manufacturing",
		"LABOURER",
		"Driver",
		"barber",
		"Lifeguard",
		"Handyman",
		"Meat Department Manager",
		"Cleaner",
		"Host Hostess LOCAL Public Eatery Garry St",
		"Loss Prevention Specialist",
		"Payroll specialist UK",
		"Sales Team Member Cotton On Westfield Bondi",
		// Not job titles at all — the board serves these as postings.
		"Page Not Found",
		"Oops something happened",
		"Dirt and worms cake",
		"YOUR JOB DESCRIPTION HERE",
		"CHECK BACK SOON",
	}

	for _, role := range roles {
		if looksTechnical(role) {
			t.Errorf("looksTechnical(%q) = true, want false", role)
		}
	}
}

// Seniority says nothing about the field. Including those words admitted
// "Senior Graphic Designer" onto a technology board.
func TestLooksTechnicalIgnoresSeniority(t *testing.T) {
	for _, role := range []string{
		"Senior Graphic Designer",
		"Junior Sales Associate",
		"Pleno Analista Comercial",
	} {
		if looksTechnical(role) {
			t.Errorf("looksTechnical(%q) = true, want false", role)
		}
	}
}

// Matching on substrings would let anything containing "ai", "go" or "dev"
// through, which is most of the English language.
func TestLooksTechnicalMatchesWholeWordsOnly(t *testing.T) {
	for _, role := range []string{
		"Retail Assistant",       // contains "ai"
		"Cargo Handler",          // contains "go"
		"Device Cleaner",         // contains "dev"
		"Maintenance Technician", // contains "ai"
		"Nodejsniz",              // not a token
	} {
		if looksTechnical(role) {
			t.Errorf("looksTechnical(%q) = true, want false — matched a substring", role)
		}
	}
}

func TestTokenizeKeepsLanguageNames(t *testing.T) {
	got := tokenize("c++ and c# developer, node.js")

	want := map[string]bool{"c++": true, "c#": true, "developer": true, "node": true, "js": true}
	for _, token := range got {
		delete(want, token)
	}
	if len(want) > 0 {
		t.Fatalf("tokenize dropped %v (got %v)", want, got)
	}
}
