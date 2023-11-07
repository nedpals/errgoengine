public class CannotBeApplied {
    public static void main(String[] args) {
        int result = addValues("5", 10); // Attempting to add a string and an integer
    }
    
    public static int addValues(int a, int b) {
        return a + b;
    }
}